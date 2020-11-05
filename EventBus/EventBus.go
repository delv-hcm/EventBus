package EventBus

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/nats-io/stan.go"
)

type IEventBus interface {
	Publish(event IntegrationEvent, topic string)
	Subscribe(event IntegrationEvent, eventHandlers []*IntegrationEventHandler, topic string)
}

type Handler struct {
	EventHandler    interface{}
	FallbackHandler interface{}
}

type EventBus struct {
	Subscriptions  map[string][]Handler
	Conn           stan.Conn
	ClientId       string
	Rb             *RingBuffer
	RetriesDurable []int
}
type ResultError struct {
	Res interface{}
	Err error
	Msg *stan.Msg
}

func (eventBus *EventBus) New() *EventBus {
	nc, err := stan.Connect("local-cluster", eventBus.ClientId, stan.NatsURL("localhost:4222"))
	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}
	return &EventBus{Subscriptions: make(map[string][]Handler), Conn: nc,
		Rb: &RingBuffer{InChan: make(chan ResultError), OutChan: make(chan ResultError, 100)}, RetriesDurable: []int{5, 10, 30}}
}

func (eventBus *EventBus) Publish(event interface{}, topic string) {
	integrationEvent := (&IntegrationEvent{}).New()
	payloadStr, _ := json.Marshal(event)
	integrationEvent.Payload = payloadStr
	integrationEvent.Type = reflect.TypeOf(event).String()

	data, err := json.Marshal(&integrationEvent)
	if err != nil {
		log.Printf("error when marshal event %v", err)
	}
	if err := eventBus.Conn.Publish(topic, data); err != nil {
		log.Printf("error when publish event %v", err)
	}
	log.Printf("Delivery message success to topic [%s]", topic)
}

func cloneEvent(evt *IntegrationEvent, topic string, delay int) {
	evt.RedeliveryCount++
	evt.Sub = fmt.Sprintf("%s-retry-%ds", topic, delay)
	evt.Delay = delay
}

func (eventBus *EventBus) publish(evt *IntegrationEvent) {
	data, err := json.Marshal(&evt)
	if err != nil {
		log.Printf("error when marshal event Id [%s] %v", evt.ID, err)
	}
	if err := eventBus.Conn.Publish(evt.Sub, data); err != nil {
		log.Printf("error when publish event Id [%s] %v", evt.ID, err)
	}
	log.Printf("Delivery event Id [%s] success to topic [%s]", evt.ID, evt.Sub)
}

func (eventBus *EventBus) prepareRetryTopic(topic string) []string {
	topics := []string{}
	for _, durationTime := range eventBus.RetriesDurable {
		topics = append(topics, fmt.Sprintf("%s-retry-%ds", topic, durationTime))
	}
	return topics
}

// QueueSubscribe
func (eventBus *EventBus) QueueSubscribe(event interface{}, eventHandlers []Handler, topic, queue string) {
	if _, ok := eventBus.Subscriptions[reflect.TypeOf(event).String()]; !ok {
		eventBus.Subscriptions[reflect.TypeOf(event).String()] = eventHandlers
	} else {
		eventBus.Subscriptions[reflect.TypeOf(event).String()] = append(eventBus.Subscriptions[reflect.TypeOf(event).String()], eventHandlers...)
	}

	go eventBus.Rb.Run(func(result *ResultError, rawMsg *stan.Msg) {
		if result.Err == nil {
			return
		} else {
			var event IntegrationEvent
			if err := json.Unmarshal(rawMsg.Data, &event); err != nil {
				log.Fatalln(err)
			}

			event.IsRetry = true
			if event.RedeliveryCount == 0 {
				cloneEvent(&event, topic, 5)
				eventBus.publish(&event)
			} else if event.RedeliveryCount == 1 {
				cloneEvent(&event, topic, 10)
				eventBus.publish(&event)
			} else if event.RedeliveryCount == 2 {
				cloneEvent(&event, topic, 30)
				eventBus.publish(&event)
			} else {
				// store message into db, manual process
				event.Sub = fmt.Sprintf("%s-fail", topic)
				eventBus.publish(&event)
			}
		}
	})

	go func() {
		_, err := eventBus.Conn.QueueSubscribe(topic, queue, func(msg *stan.Msg) {
			var data IntegrationEvent
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Printf("Error when Unmarshal event: %v", err)
			} else {
				for _, handler := range eventBus.Subscriptions[data.Type] {
					eventHandler := handler.EventHandler
					concrete := reflect.ValueOf(eventHandler)
					go func() {
						log.Printf("[Queue] Handle event Id [%s] redelivery [%d] from topic [%s] with EventHandler [%v]", data.ID,
							data.RedeliveryCount, topic, reflect.TypeOf(eventHandler).String())
						concrete.MethodByName("Handle").Call([]reflect.Value{reflect.ValueOf(msg), reflect.ValueOf(eventBus.Rb.InChan)})
					}()
				}
			}
		}, stan.DurableName(queue))
		if err != nil {
			log.Printf("error when subscribe topic [%s]. Error --> %v", topic, err)
		}
	}()

	retriesTopic := eventBus.prepareRetryTopic(topic)
	for _, retryTopic := range retriesTopic {
		go func(topic, queue string) {
			_, err := eventBus.Conn.QueueSubscribe(topic, queue, func(msg *stan.Msg) {
				var data IntegrationEvent
				if err := json.Unmarshal(msg.Data, &data); err != nil {
					log.Printf("Error when Unmarshal event: %v", err)
				} else {
					for _, handler := range eventBus.Subscriptions[data.Type] {
						item := handler.FallbackHandler
						concrete := reflect.ValueOf(item)
						go func() {
							log.Printf("[Queue] Handle event Id [%s] redelivery [%d] from topic [%s] with FallbackHandler [%v]", data.ID,
								data.RedeliveryCount, data.Sub, reflect.TypeOf(item).String())
							if data.IsRetry {
								log.Printf("Sleeping in %d seconds", data.Delay)
								time.Sleep(time.Duration(data.Delay) * time.Second)
							}
							concrete.MethodByName("Handle").Call([]reflect.Value{reflect.ValueOf(msg), reflect.ValueOf(eventBus.Rb.InChan)})
						}()
					}
				}
			}, stan.DurableName(queue))
			if err != nil {
				log.Printf("error when subscribe retry topic [%s]. Error --> %v", topic, err)
			}
		}(retryTopic, "queue-retry")
	}

}
