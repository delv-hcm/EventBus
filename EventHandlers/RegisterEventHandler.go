package EventHandlers

import (
	"encoding/json"
	"eventbus/EventBus"
	"log"
	"time"
)

type RegisterEvent struct {
	Event    *EventBus.Event
	ClassID  string
	SchoolID string
}

type RegisterEventHandler struct {
}

func (handler *RegisterEventHandler) Handle(event []byte, errorResult chan<- EventBus.ResultError) {
	var evt RegisterEvent
	err := json.Unmarshal(event, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{}
	}
	time.Sleep(time.Duration(1) * time.Microsecond)
	log.Println("registerEventHandler classId:", evt.ClassID)
	errorResult <- EventBus.ResultError{Err: nil}
}

type RegisterEvent2Handler struct {
}

func (handler *RegisterEvent2Handler) Handle(event []byte, errorResult chan<- EventBus.ResultError) {
	var evt RegisterEvent
	err := json.Unmarshal(event, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{}
	}
	time.Sleep(time.Duration(350) * time.Microsecond)
	log.Println("registerEventHandler2 classId:", evt.ClassID)
	errorResult <- EventBus.ResultError{Err: nil}
}
