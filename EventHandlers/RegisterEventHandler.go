package EventHandlers

import (
	"encoding/json"
	"errors"
	"eventbus/EventBus"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/nats-io/stan.go"
)

type RegisterEvent struct {
	Event    *EventBus.Event
	ClassID  string
	SchoolID string
}

type RegisterEventHandler struct {
}

func (handler *RegisterEventHandler) Handle(msg *stan.Msg, errorResult chan<- EventBus.ResultError) {
	var evt RegisterEvent
	err := json.Unmarshal(msg.Data, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{Res: nil, Err: err, Seq: msg.Sequence}
	}
	if evt.ClassID == "2" {
		errorResult <- EventBus.ResultError{Res: nil, Err: errors.New("Unexpected error"), Seq: msg.Sequence, Msg: msg}
		return
	}
	// handle business logic
	time.Sleep(time.Duration(rand.Intn(100)*1) * time.Microsecond)
	log.Printf("NumGoroutine [%d] Invoke [RegisterEventHandler], classId: %s", runtime.NumGoroutine(), evt.ClassID)
	// return
	errorResult <- EventBus.ResultError{Res: fmt.Sprintf("done RegisterEventHandler: %s", evt.ClassID), Err: nil}
}

type RegisterEvent2Handler struct {
}

func (handler *RegisterEvent2Handler) Handle(msg *stan.Msg, errorResult chan<- EventBus.ResultError) {
	var evt RegisterEvent
	err := json.Unmarshal(msg.Data, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{Res: nil, Err: err}
	}
	// handle business logic
	time.Sleep(time.Duration(rand.Intn(500)*1) * time.Microsecond)
	log.Printf("NumGoroutine [%d] Invoke [RegisterEventHandler2], classId: %s", runtime.NumGoroutine(), evt.ClassID)
	// return
	errorResult <- EventBus.ResultError{Res: fmt.Sprintf("done RegisterEventHandler2: %s", evt.ClassID), Err: nil}
}
