package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-redis/redis"
	db "github.com/rresender/csta-integration/sample/common"
	"github.com/streadway/amqp"
	"github.com/tidwall/gjson"
)

var (
	conn   *redis.Client
	mq     *amqp.Connection
	topics []Topic
)

// Agent Object
type Agent struct {
	ID      string
	Station string
	Skills  []string
}

// Call Object
type Call struct {
	UCID         string
	UUI          string
	AgentStation string
	AgentID      string
	Skill        string
	VDN          string
	ANI          string
}

// Topic struct
type Topic struct {
	Name string
	Type string
}

func init() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis:6379"
	}

	conn = db.Connect(redisHost)

	rabbitMQHost := os.Getenv("RABBITMQ_PORT_5672_TCP_ADDR")
	if redisHost == "" {
		rabbitMQHost = ":5672"
	}
	var err error
	mq, err = amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:5672/", rabbitMQHost))

	db.FailOnError(err, "Failed to connect to RabbitMQ")
	log.Println("RabbitMQ connected...")

	extensionsAndTypes := strings.Split(os.Getenv("MONITORED_EXTENSIONS"), ",")
	if len(extensionsAndTypes) > 1 {
		for i := range extensionsAndTypes {
			extensionAndType := strings.Split(extensionsAndTypes[i], ":")
			name := strings.TrimSpace(extensionAndType[0])
			extType := strings.TrimSpace(extensionAndType[1])
			topics = append(topics, Topic{Name: name, Type: extType})
		}
	}
}

func deliveredEvent(event string) bool {
	return strings.HasPrefix(event, "{\"DeliveredEvent\"")
}

func estabilishedEvent(event string) bool {
	return strings.HasPrefix(event, "{\"EstablishedEvent\"")
}

func agentLoggedOffEvent(event string) bool {
	return strings.HasPrefix(event, "{\"AgentLoggedOffEvent\"")
}

func agentLoggedOnEvent(event string) bool {
	return strings.HasPrefix(event, "{\"AgentLoggedOnEvent\"")
}

func getGeliveredEventValue(path string, json string) string {
	return gjson.Get(json, "DeliveredEvent."+path).String()
}

func getEstabilishedEventValue(path string, json string) string {
	return gjson.Get(json, "EstablishedEvent."+path).String()
}

func getExtensionNumber(path string, json string) string {
	return strings.Split(gjson.Get(json, path+".deviceIdentifier.#content").String(), ":")[0]
}

func getUCID(eventType string, json string) string {
	return gjson.Get(json, eventType+".callLinkageData.globalCallData.globalCallLinkageID.globallyUniqueCallLinkageID").String()
}

func save(ID string, entity interface{}) {
	value, err := json.Marshal(entity)
	if err != nil {
		log.Printf("error while marshaling json %v\n", err)
	}
	if err := conn.Set(ID, value, 0).Err(); err != nil {
		log.Printf("error: %v", err)
	}
}

func find(ID string, entity interface{}) error {
	v, err := conn.Get(ID).Bytes()
	if err != nil {
		return err
	}
	if v == nil || len(v) <= 1 {
		return fmt.Errorf("entity could not be found for ID %s", ID)
	}
	if err := json.Unmarshal(v, &entity); err != nil {
		log.Printf("error while marshaling json %v\n", err)
		return err
	}
	return nil
}

func remove(ID string) error {
	if err := conn.Del(ID).Err(); err != nil {
		return err
	}
	return nil
}

func addCall(event string) {
	call := Call{
		UCID: getUCID("DeliveredEvent", event),
		VDN:  getExtensionNumber("DeliveredEvent.calledDevice", event),
		ANI:  getExtensionNumber("DeliveredEvent.callingDevice", event)}
	save(call.UCID, &call)
	log.Printf("Call saved %v\n", call)
}

func getAgentIDKey(station string) string {
	return "agent-station-" + station
}

func updateCall(event string) {
	var call Call
	UCID := getUCID("EstablishedEvent", event)
	if err := find(UCID, &call); err != nil {
		log.Printf("%v error", err)
		return
	}
	UUI, _ := hex.DecodeString(getEstabilishedEventValue("userData.string", event))
	call.UUI = string(UUI)
	call.AgentStation = getExtensionNumber("EstablishedEvent.answeringDevice", event)
	call.Skill = getSkill(event)
	var agent Agent
	find(getAgentIDKey(call.AgentStation), &agent)
	call.AgentID = agent.ID
	save(call.UCID, &call)
	log.Printf("Call updated %v\n", call)
}

func getSkill(event string) string {
	return strings.Split(getEstabilishedEventValue("extensions.privateData.private.EstablishedEventPrivateData.acdGroup.#content", event), ":")[0]
}

func addAgent(event string) {
	station := strings.Split(gjson.Get(event, "AgentLoggedOnEvent.agentDevice.deviceIdentifier.#content").String(), ":")[0]
	skill := strings.Split(gjson.Get(event, "AgentLoggedOnEvent.acdGroup.#content").String(), ":")[0]
	var agent Agent
	if err := find(getAgentIDKey(station), &agent); err != nil {
		agent =
			Agent{
				ID:      gjson.Get(event, "AgentLoggedOnEvent.agentID").String(),
				Station: station,
				Skills:  []string{skill},
			}
	} else {
		agent.Skills = append(agent.Skills, skill)
	}
	save(getAgentIDKey(agent.Station), agent)
	log.Printf("Agent added %v\n", agent)
}

func removeAgent(event string) {
	station := strings.Split(gjson.Get(event, "AgentLoggedOffEvent.agentDevice.deviceIdentifier.#content").String(), ":")[0]
	ID := gjson.Get(event, "AgentLoggedOffEvent.agentID").String()
	remove(getAgentIDKey(station))
	log.Printf("Agent removed %s\n", ID)
}

func createConsumer(queue string) (<-chan amqp.Delivery, *amqp.Channel, error) {

	ch, err := mq.Channel()
	db.FailOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		queue,    // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	db.FailOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	db.FailOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		queue,  // exchange
		false,
		nil)
	db.FailOnError(err, "Failed to bind a queue")

	events, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	db.FailOnError(err, "Failed to register a consumer")

	return events, ch, err
}

func monitoringVDN(vdn string) (*amqp.Channel, error) {
	events, channel, err := createConsumer(vdn)

	go func() {
		for e := range events {
			event := string(e.Body)
			switch {
			case deliveredEvent(event):
				addCall(event)
			case estabilishedEvent(event):
				updateCall(event)
			}
		}
	}()

	return channel, err
}

func monitoringSkill(skill string) error {
	events, channel, err := createConsumer(skill)
	go func() {
		defer channel.Close()
		for e := range events {
			event := string(e.Body)
			switch {
			case agentLoggedOnEvent(event):
				addAgent(event)
			case agentLoggedOffEvent(event):
				removeAgent(event)
			}
		}
	}()
	return err
}

// MessageHandler MessageHandler
func messageHandler(topics []Topic) {

	for _, t := range topics {
		log.Printf("Handling messages for %v\n", t)
		switch t.Type {
		case "SKILL":
			monitoringSkill(t.Name)
		case "VDN":
			monitoringVDN(t.Name)
		}
	}
}

func main() {

	defer Close()

	loop := make(chan bool)

	messageHandler(topics)

	log.Printf(" [*] Waiting for messages...")
	<-loop
}

// Close method
func Close() {
	defer conn.Close()
	defer mq.Close()
}
