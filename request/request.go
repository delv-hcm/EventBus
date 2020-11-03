package main

import (
	"eventbus/EventBus"
	"eventbus/EventHandlers"
	"strconv"
)

func main() {
	eventBus := (&EventBus.EventBus{ClientId: "client-01"}).New()
	index := 0
	for range []int{0, 1, 2, 4} {
		eventBus.Publish(&EventHandlers.RegisterEvent{ClassID: strconv.Itoa(index)}, "register-event-5")
		index++
	}
}
