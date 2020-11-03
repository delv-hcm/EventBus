package EventBus

import "log"

type Event struct {
	ID      string
	Payload interface{}
}

type EventHandler struct {
}

func (eventHandler *EventHandler) Handle(event *Event, errorResult chan<- ResultError) {
	log.Println(event)
	errorResult <- ResultError{Res: "success", Err: nil}
}
