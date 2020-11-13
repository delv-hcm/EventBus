package EventHandlers

import (
	"encoding/json"
	"errors"
	"eventbus/EventBus"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"time"

	"github.com/nats-io/stan.go"
)

type RegisterEvent struct {
	ClassID  string
	SchoolID string
}

type RegisterEventHandler struct {
}

func (handler *RegisterEventHandler) New() *RegisterEventHandler {
	return &RegisterEventHandler{}
}

func (handler *RegisterEventHandler) Handle(msg *stan.Msg, errorResult chan<- EventBus.ResultError) {
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
	// if data.ClassID == "1" {
	// 	errorResult <- EventBus.ResultError{Res: nil, Err: errors.New("Unexpected error"), Msg: msg}
	// 	return
	// }
	time.Sleep(time.Duration(rand.Intn(100)*10000) * time.Microsecond)
	log.Printf("NumGoroutine [%d] Invoke [RegisterEventHandler], classId: %s, eventId: %s", runtime.NumGoroutine(), data.ClassID, evt.ID)
	// return
	errorResult <- EventBus.ResultError{Res: fmt.Sprintf("done RegisterEventHandler: %s", data.ClassID), Err: nil}
}

type RegisterEvent2Handler struct {
}

func (handler *RegisterEvent2Handler) Handle(msg *stan.Msg, errorResult chan<- EventBus.ResultError) {
	var evt EventBus.IntegrationEvent
	err := json.Unmarshal(msg.Data, &evt)
	if err != nil {
		errorResult <- EventBus.ResultError{Res: nil, Err: err, Msg: msg}
	}
	var data RegisterEvent
	if err := json.Unmarshal(evt.Payload, &data); err != nil {
		errorResult <- EventBus.ResultError{Res: nil, Err: errors.New("Json Unexpected error"), Msg: msg}
		return
	}
	if data.ClassID == "2" {
		errorResult <- EventBus.ResultError{Res: nil, Err: errors.New("Unexpected error"), Msg: msg, HandlerType: reflect.TypeOf(handler).String()}
		return
	}
	// handle business logic
	time.Sleep(time.Duration(rand.Intn(500)*10000) * time.Microsecond)
	log.Printf("NumGoroutine [%d] Invoke [RegisterEventHandler2], classId: %s, eventId: %s", runtime.NumGoroutine(), data.ClassID, evt.ID)
	// return
	errorResult <- EventBus.ResultError{Res: fmt.Sprintf("done RegisterEventHandler2: %s", data.ClassID), Err: nil}
}
