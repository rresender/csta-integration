package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rresender/csta-integration/cti/db"
	"github.com/rresender/csta-integration/cti/helper"
	"github.com/rresender/csta-integration/cti/provider"
	"github.com/rresender/csta-integration/cti/rabbitmq"
	"github.com/rresender/csta-integration/cti/redis"

	xj "github.com/basgys/goxml2json"
	"github.com/gorilla/mux"
)

var (
	applicationName string
	provider_host   string
	pbx             string
	user            string
	password        string
	extensionsMap   map[string]*db.Extension
)

// UnsolicitedInvokeID generic InvokeID for unsolicited events
var UnsolicitedInvokeID = strconv.Itoa(db.MaxInvokeID)

func init() {
	applicationName = "provider-monitoring-" + helper.GetLocalIP()

	provider_host = os.Getenv("PROVIDER_HOST")
	if provider_host == "" {
		provider_host = "localhost:4721"
	}
	pbx = os.Getenv("PBX_HOST")
	if pbx == "" {
		pbx = "localhost"
	}
	user = os.Getenv("CTI_USER")
	if user == "" {
		user = "ctiuser"
	}
	password = os.Getenv("CTI_PASSWORD")
	if password == "" {
		password = "ctipassword"
	}
	extensionsAndTypes := strings.Split(os.Getenv("MONITORED_EXTENSIONS"), ",")
	if len(extensionsAndTypes) > 1 {
		extensionsMap = make(map[string]*db.Extension)
		for i := range extensionsAndTypes {
			extensionAndType := strings.Split(extensionsAndTypes[i], ":")
			ID := strings.TrimSpace(extensionAndType[0])
			extType := strings.TrimSpace(extensionAndType[1])
			extensionsMap[ID] = &db.Extension{ID: ID, Type: extType}
		}
	}
}

//Handler event listener
type Handler struct{}

func readResponse(invokeID string, appName string) string {
	responseDataReceivedTicker := time.Tick(time.Second * 2)
	for {
		select {
		case <-responseDataReceivedTicker:
			data := db.Find(helper.GetInvokeIDKey(invokeID, appName))
			if data == "" {
				continue
			}
			return data
		}
	}
}

func heartbeat(sessionID string, appName string) {
	go func() {
		heartbeatTicker := time.Tick(time.Second * 30)
		invokeID := db.GetInvoke("heartbeat", appName)
		for {
			select {
			case <-heartbeatTicker:
				provider.Send(invokeID, provider.ResetApplicationSessionTimerMessage(sessionID))
				data := readResponse(invokeID, appName)
				var response provider.ResetApplicationSessionTimerResponse
				provider.ParseMessageResponse(data, &response)
				if err := provider.ParseMessageResponse(data, &response); err != nil {
					log.Println(err)
				}
				db.Delete(helper.GetQueueKey(invokeID, appName))
			}
		}
	}()
}

// DoProcess process responses from provider_host
func (h Handler) DoProcess(invokeID string, data string) {
	switch invokeID {
	case UnsolicitedInvokeID:
		converted, _ := xj.Convert(bytes.NewBufferString(data))
		var in map[string]interface{}
		json.Unmarshal(converted.Bytes(), &in)
		out := make(map[string]string)
		helper.ParseJSON(in, out)
		monitorCrossRefID := out["monitorCrossRefID"]
		queue := db.Find(helper.GetMonitorCrossRefIDKey(monitorCrossRefID, applicationName))
		rabbitmq.Send(queue, converted)
	default:
		db.SaveWithTTL(helper.GetInvokeIDKey(invokeID, applicationName), data)
	}
}

func stopMonitoring(extension string, appName string) (*db.Extension, error) {
	ext := db.FindExtension(extension)
	if ext == nil {
		return nil, errors.New("extension could not be found")
	}
	invokeID := db.GetInvoke(extension, appName)

	provider.Send(invokeID, provider.MonitorStopMessage(ext.MonitorCrossRefID))
	data := readResponse(invokeID, appName)

	var response provider.MonitorStopResponse
	var err error
	if err = provider.ParseMessageResponse(data, &response); err != nil {
		log.Println(err)
	}

	db.RemoveExtensionFromList(extension)
	db.Delete(extension)
	db.Delete(helper.GetMonitorCrossRefIDKey(ext.MonitorCrossRefID, appName))
	delete(extensionsMap, ext.ID)
	return ext, err
}

func cleanUpHook(sessionID string, appName string) {

	defer provider.Close()
	defer redis.Close()
	defer rabbitmq.Close()

	err := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)
		err <- fmt.Errorf("signal %v", <-c)
	}()

	log.Println("Cleaning up... :", <-err)

	for _, extension := range extensionsMap {
		stopMonitoring(extension.ID, appName)
	}

	if sessionID != "" {
		invokeID := db.GetInvoke(appName, appName)
		provider.Send(invokeID, provider.StopAppSessionMessage(sessionID))
		data := readResponse(invokeID, appName)
		var response provider.StartApplicationSessionResponse
		if err := provider.ParseMessageResponse(data, &response); err != nil {
			log.Println(err)
		}
	}
}

func startSession(appName string, user string, password string, sessionCleanupDelay string, requestedSessionDuration string) (string, error) {

	invokeID := db.GetInvoke(appName, appName)

	provider.Send(invokeID, provider.StartApplicationSessionMessage(appName, user, password, sessionCleanupDelay, requestedSessionDuration))
	session := readResponse(invokeID, appName)

	var response provider.StartApplicationSessionResponse
	if err := provider.ParseMessageResponse(session, &response); err != nil {
		return "", err
	}
	log.Printf("SessionID: %s\n", response.SessionID)
	return response.SessionID, nil
}

func getDeviceID(extension string, pbx string, appName string) (string, error) {

	invokeID := db.GetInvoke(extension, appName)

	provider.Send(invokeID, provider.GetDeviceIDMessage(pbx, extension))
	device := readResponse(invokeID, appName)

	var response provider.GetDeviceIDResponse
	if err := provider.ParseMessageResponse(device, &response); err != nil {
		return "", err
	}
	log.Printf("DeviceID: %s\n", response.Device.ID)

	return response.Device.ID, nil
}

func getMonitorCrossRefID(invokeID string, extension string, appName string) (string, error) {

	monitorCrossRefID := readResponse(invokeID, appName)

	var response provider.MonitorStartResponse
	if err := provider.ParseMessageResponse(monitorCrossRefID, &response); err != nil {
		return "", err
	}
	log.Printf("MonitorCrossRefID: %s\n", response.MonitorCrossRefID)

	db.Save(helper.GetMonitorCrossRefIDKey(response.MonitorCrossRefID, appName), extension)
	return response.MonitorCrossRefID, nil
}

func startVDNMonitoring(extension string, pbx string, appName string) (*db.Extension, error) {
	deviceID, err := getDeviceID(extension, pbx, appName)
	if err != nil {
		return nil, err
	}
	invokeID := db.GetInvoke(extension, appName)
	provider.Send(invokeID, provider.MonitorVDNStartMessage(deviceID))
	monitorCrossRefID, err := getMonitorCrossRefID(invokeID, extension, appName)
	return &db.Extension{ID: extension, Type: "VDN", DeviceID: deviceID, MonitorCrossRefID: monitorCrossRefID}, err
}

func startSkillMonitoring(extension string, pbx string, appName string) (*db.Extension, error) {
	deviceID, err := getDeviceID(extension, pbx, appName)
	if err != nil {
		return nil, err
	}
	invokeID := db.GetInvoke(extension, appName)
	provider.Send(invokeID, provider.MonitorSkillStartMessage(deviceID))
	monitorCrossRefID, err := getMonitorCrossRefID(invokeID, extension, appName)
	return &db.Extension{ID: extension, Type: "SKILL", DeviceID: deviceID, MonitorCrossRefID: monitorCrossRefID}, err
}

func doMonitoring(extension string, extType string, appName string) (*db.Extension, error) {
	if db.Exists(extension) {
		log.Printf("the extension %s is already being monitored. The monitoring process will be restarted...\n", extension)
		stopMonitoring(extension, appName)
	}
	var ext *db.Extension
	var err error
	extType = strings.ToUpper(extType)
	switch extType {
	case "VDN":
		ext, err = startVDNMonitoring(extension, pbx, appName)
	case "SKILL":
		ext, err = startSkillMonitoring(extension, pbx, appName)
	default:
		err = fmt.Errorf("type %s is not valid", extType)
	}
	if err == nil {
		db.AddExtensionToList(ext.ID)
		db.SaveExtension(ext)
		extensionsMap[ext.ID] = ext
	}
	return ext, err
}

func httpHandler(pbx string, appName string) {
	go func() {
		m := mux.NewRouter()
		m.HandleFunc("/start/{type}/{extension}", func(w http.ResponseWriter, r *http.Request) {

			vars := mux.Vars(r)
			extension := vars["extension"]
			extType := vars["type"]

			ext, err := doMonitoring(extension, extType, appName)

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				rabbitmq.DeleteQueue(extension)
				return
			}

			fmt.Fprintf(w, "Monitoring on %s: %s (MonitorCrossRefID: %s) has been started", extType, extension, ext.MonitorCrossRefID)
		})

		m.HandleFunc("/stop/{extension}", func(w http.ResponseWriter, r *http.Request) {

			vars := mux.Vars(r)
			extension := vars["extension"]

			ext, err := stopMonitoring(extension, appName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if ext == nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "extension: %s has not been monitored", extension)
				return
			}

			fmt.Fprintf(w, "Monitoring on %s (MonitorCrossRefID: %s) has been stopped", extension, ext.MonitorCrossRefID)

		})

		m.HandleFunc("/getall", func(w http.ResponseWriter, r *http.Request) {
			extensions := db.GetAllExtensions()
			fmt.Fprintf(w, "List of extensions being monitored: %d\n", len(extensions))
			for _, extension := range extensions {
				fmt.Fprintln(w, extension)
			}
		})

		log.Fatal(http.ListenAndServe(":7700", m))
	}()
}

func monitoringExtensions(appName string) {
	go func() {
		for ID, elem := range extensionsMap {
			ext, err := doMonitoring(ID, elem.Type, appName)
			if err != nil {
				log.Printf("%v\n", err)
				rabbitmq.DeleteQueue(ID)
				return
			}
			log.Printf("Monitoring on %s: %s (MonitorCrossRefID: %s) has been started\n", ext.Type, ext.ID, ext.MonitorCrossRefID)
		}
	}()
}

func main() {

	provider.Connect(provider_host, &Handler{})

	sessionID, err := startSession(applicationName, user, password, "60", "180")
	if err != nil {
		log.Fatalln(err)
	}

	monitoringExtensions(applicationName)

	heartbeat(sessionID, applicationName)

	httpHandler(pbx, applicationName)

	cleanUpHook(sessionID, applicationName)
}
