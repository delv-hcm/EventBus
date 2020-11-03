package EventHandlers

import (
	"encoding/json"
	"eventbus/EventBus"
	"log"
	"runtime"
	"time"
)

type LeaveClassEvent struct {
	Event   *EventBus.Event
	ClassID string
	IsLeave bool
}

type LeaveClassEventHandler struct {
}

func (handler *LeaveClassEventHandler) Handle(event []byte, errorResult chan<- EventBus.ResultError) {
	var evt LeaveClassEvent
	err := json.Unmarshal(event, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{}
	}
	log.Println("in handler", runtime.NumGoroutine())
	time.Sleep(time.Duration(1) * time.Second)
	log.Println("isleave:", evt.IsLeave)
	errorResult <- EventBus.ResultError{Err: nil}
}
