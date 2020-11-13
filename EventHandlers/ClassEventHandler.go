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

type ClassEvent struct {
	ClassID string
}

type ClassEventHandler struct {
}

func (handler *ClassEventHandler) Handle(msg *stan.Msg, errorResult chan<- EventBus.ResultError) {
	var evt EventBus.IntegrationEvent
	err := json.Unmarshal(msg.Data, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{Res: nil, Err: err, Msg: msg}
	}
	// handle business logic
	var data RegisterEvent
	if err := json.Unmarshal(evt.Payload, &data); err != nil {
		errorResult <- EventBus.ResultError{Res: nil, Err: errors.New("Json Unexpected error"), Msg: msg}
		return
	}
	if data.ClassID == "3" {
		errorResult <- EventBus.ResultError{Res: nil, Err: errors.New("Unexpected error"), Msg: msg}
		return
	}
	time.Sleep(time.Duration(rand.Intn(500)*10000) * time.Microsecond)
	log.Printf("NumGoroutine [%d] Invoke [ClassEventHandler], classId: %s, eventId: %s", runtime.NumGoroutine(), data.ClassID, evt.ID)

	errorResult <- EventBus.ResultError{Res: fmt.Sprintf("done ClassEventHandler: %s", data.ClassID), Err: nil}
}
