package EventHandlers

import (
	"encoding/json"
	"eventbus/EventBus"
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
	time.Sleep(time.Duration(1) * time.Second)
	errorResult <- EventBus.ResultError{Err: nil}
}
