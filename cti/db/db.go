package db

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"

	"github.com/rresender/csta-integration/cti/helper"
	"github.com/rresender/csta-integration/cti/redis"
)

// MaxInvokeID allowed
var MaxInvokeID = 9999

//Extension object
type Extension struct {
	ID                string
	Type              string
	DeviceID          string
	MonitorCrossRefID string
}

// SaveWithTTL data with a time-to-live
func SaveWithTTL(key string, value string) {
	redis.Set(key, []byte(value))
	redis.Expire(key)
}

//Save data
func Save(key string, value string) {
	redis.Set(key, []byte(value))
}

//Find data
func Find(key string) string {
	v, _ := redis.Get(key)
	return string(v)
}

// Delete an entry
func Delete(key string) {
	redis.Delete(key)
}

//Exists data
func Exists(key string) bool {
	e, _ := redis.Exists(key)
	return e
}

func incrementInvoke(appName string) string {
	var l sync.Mutex
	l.Lock()
	defer l.Unlock()
	ID := appName + "invokeID"
	c, _ := redis.Incr(ID)
	if c >= MaxInvokeID {
		redis.Set(ID, []byte("0"))
		c, _ = redis.Incr(ID)
	}
	invokeID := helper.Lpad(strconv.Itoa(c), "0", 4)
	return invokeID
}

// GetInvoke for processing requests
func GetInvoke(queueID string, appName string) string {
	invokeID := incrementInvoke(appName)
	SaveWithTTL(helper.GetQueueKey(invokeID, appName), queueID)
	return invokeID
}

// GetInvokeNoTTL for processing requests without time-to-live
func GetInvokeNoTTL(queueID string, appName string) string {
	invokeID := incrementInvoke(appName)
	Save(helper.GetQueueKey(invokeID, appName), queueID)
	return invokeID
}

// GetAllExtensions being managed
func GetAllExtensions() []string {
	return redis.GetValues("extensions")
}

// AddExtensionToList to the memory
func AddExtensionToList(extension string) {
	redis.PushValue("extensions", extension)
}

// RemoveExtensionFromList from the memory
func RemoveExtensionFromList(extension string) {
	redis.RemoveValue("extensions", extension)
}

//FindExtension Object
func FindExtension(ID string) *Extension {
	v, _ := redis.Get(ID)
	if v == nil {
		return nil
	}
	var extension Extension
	if err := json.Unmarshal(v, &extension); err != nil {
		log.Printf("error while marshaling json %v\n", err)
		return nil
	}
	return &extension
}

//SaveExtension Object
func SaveExtension(extension *Extension) {
	value, err := json.Marshal(&extension)
	if err != nil {
		log.Printf("error while marshaling json %v\n", err)
	}
	redis.Set(extension.ID, value)
}
