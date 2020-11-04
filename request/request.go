package main

import (
	"eventbus/EventBus"
	"eventbus/EventHandlers"
	"math/rand"
	"strconv"
	"time"
)

func main() {
	eventBus := (&EventBus.EventBus{ClientId: "client-01"}).New(10)
	//index := 0

	eventBus.Publish(&EventHandlers.RegisterEvent{ClassID: strconv.Itoa(2)}, "register-event-5")
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
}
