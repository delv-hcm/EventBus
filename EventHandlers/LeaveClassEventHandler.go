package EventHandlers

import (
	"encoding/json"
	"eventbus/EventBus"
	"log"
	"runtime"
	"time"

	"github.com/nats-io/stan.go"
)

type LeaveClassEvent struct {
	Event   *EventBus.Event
	ClassID string
	IsLeave bool
}

type LeaveClassEventHandler struct {
}

func (handler *LeaveClassEventHandler) Handle(msg *stan.Msg, errorResult chan<- EventBus.ResultError) {
	var evt LeaveClassEvent
	err := json.Unmarshal(msg.Data, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{Res: nil, Err: err}
		return
	}
	// handle business logic
	time.Sleep(time.Duration(350) * time.Microsecond)
	log.Printf("NumGoroutine [%d] Invoke [LeaveClassEventHandler], classId: %s", runtime.NumGoroutine(), evt.ClassID)
	// return
	errorResult <- EventBus.ResultError{Res: "success leave", Err: nil}
}
