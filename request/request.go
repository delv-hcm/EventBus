package main

import (
	"math/rand"
	"strconv"
	"time"

	"eventbus/EventBus"
	"eventbus/EventHandlers"
)

func main() {
	eventBus := (&EventBus.EventBus{ClientId: "client-01"}).New(10)
	index := 0
	for {
		eventBus.Publish(&EventHandlers.RegisterEvent{ClassID: strconv.Itoa(index)}, "register-event-5")
		time.Sleep(time.Duration(rand.Intn(500)*1000) * time.Microsecond)
		eventBus.Publish(&EventHandlers.LeaveClassEvent{ClassID: strconv.Itoa(index)}, "leave-class-event")
		time.Sleep(time.Duration(rand.Intn(250)*2000) * time.Microsecond)
		index++
	}
}
