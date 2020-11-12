package main

import (
	"math/rand"
	"strconv"
	"time"

	"eventbus/EventBus"
	"eventbus/EventHandlers"
)

func main() {
	eventBus := (&EventBus.EventBus{ClientId: "client-01"}).New()
	index := 0
	for {
		if index == 3 {
			break
		}
		eventBus.Publish(&EventHandlers.RegisterEvent{ClassID: strconv.Itoa(index)}, "register-event-100")
		time.Sleep(time.Duration(rand.Intn(500)*100) * time.Microsecond)
		//	eventBus.Publish(&EventHandlers.LeaveClassEvent{ClassID: strconv.Itoa(index)}, "leave-class-event")
		//	time.Sleep(time.Duration(rand.Intn(250)*2000) * time.Microsecond)
		//	eventBus.Publish(&EventHandlers.ClassEvent{ClassID: strconv.Itoa(index)}, "class_event")
		index++
	}
}
