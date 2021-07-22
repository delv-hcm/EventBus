package EventBus

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/nats-io/stan.go"
)

const (
	cBlack = "\u001b[90m"
	cRed   = "\u001b[91m"
	cCyan  = "\u001b[96m"
	// cGreen = "\u001b[92m"
	cYellow = "\u001b[93m"
	cBlue   = "\u001b[94m"
	// cMagenta = "\u001b[95m"
	// cWhite   = "\u001b[97m"
	cReset = "\u001b[0m"
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
	Rw             sync.Mutex
}
type ResultError struct {
	Res         interface{}
	Err         error
	Msg         *stan.Msg
	HandlerType string
}

func (eventBus *EventBus) Run() {
	go eventBus.Rb.Run(func(result *ResultError, rawMsg *stan.Msg) {
		if result.Err == nil {
			return
		} else {
			var event IntegrationEvent
			if err := json.Unmarshal(rawMsg.Data, &event); err != nil {
				log.Fatalln(err)
			}

			event.HandleBy = result.HandlerType

			if event.RedeliveryCount == 0 {
				evt := eventBus.cloneEvent(event, event.OriginSub, 5)
				eventBus.publish(evt)
			} else if event.RedeliveryCount == 1 {
				evt := eventBus.cloneEvent(event, event.OriginSub, 10)
				eventBus.publish(evt)
			} else if event.RedeliveryCount == 2 {
				evt := eventBus.cloneEvent(event, event.OriginSub, 30)
				eventBus.publish(evt)
			} else {
				evt := eventBus.cloneEvent(event, event.OriginSub, -1)
				eventBus.publish(evt)
			}
		}
	})
}

func (eventBus *EventBus) New() *EventBus {
	nc, err := stan.Connect("local-cluster", eventBus.ClientId, stan.NatsURL("localhost:4222"))
	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}

	bus := &EventBus{Subscriptions: make(map[string][]Handler, 10), Conn: nc,
		Rb: &RingBuffer{InChan: make(chan ResultError), OutChan: make(chan ResultError, 100)}, RetriesDurable: []int{5, 10, 30}}

	bus.Run()

	return bus
}

func (eventBus *EventBus) Publish(event interface{}, topic string) {
	integrationEvent := (&IntegrationEvent{}).New()
	payloadStr, _ := json.Marshal(event)
	integrationEvent.Payload = payloadStr
	integrationEvent.Type = reflect.TypeOf(event).String()
	integrationEvent.Sub = topic
	integrationEvent.OriginSub = topic

	data, err := json.Marshal(&integrationEvent)
	if err != nil {
		log.Printf("error when marshal event %v", err)
	}
	if err := eventBus.Conn.Publish(topic, data); err != nil {
		log.Printf("error when publish event %v", err)
	}
	log.Printf("Delivery message success to topic [%s]", topic)
}

func (eventBus *EventBus) cloneEvent(evt IntegrationEvent, topic string, delay int) IntegrationEvent {
	if delay == -1 {
		evt.IsRetry = false
		evt.Sub = fmt.Sprintf("%s-%s", evt.OriginSub, "fail")
		evt.Delay = delay
	} else {
		evt.IsRetry = true
		evt.RedeliveryCount += 1
		evt.Sub = fmt.Sprintf("%s-retry-after%ds", topic, delay)
		evt.Delay = delay
	}

	return evt
}

func (eventBus *EventBus) publish(evt IntegrationEvent) {
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("error when marshal event Id [%s] %v", evt.ID, err)
	}
	if err := eventBus.Conn.Publish(evt.Sub, data); err != nil {
		log.Printf("error when publish event Id [%s] %v", evt.ID, err)
	}
	log.Printf("Delivery event Id [%s] success to topic [%s%s%s]", evt.ID, cCyan, evt.Sub, cReset)
}

func (eventBus *EventBus) prepareRetryTopic(topic string) []string {
	topics := []string{}
	for _, durationTime := range eventBus.RetriesDurable {
		topics = append(topics, fmt.Sprintf("%s-retry-after%ds", topic, durationTime))
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

	go func() {
		_, err := eventBus.Conn.QueueSubscribe(topic, queue, func(msg *stan.Msg) {
			var data IntegrationEvent
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				log.Printf("Error when Unmarshal event: %v", err)
			} else {
				for _, handler := range eventBus.Subscriptions[data.Type] {
					eventHandler := handler.EventHandler
					concrete := reflect.ValueOf(eventHandler)
					log.Printf("%s[Queue]%s Handle event Id [%s] redelivery [%d] from topic [%s%s%s] with EventHandler [%s%v%s]", cBlue, cReset, data.ID,
						data.RedeliveryCount, cCyan, topic, cReset, cCyan, reflect.TypeOf(eventHandler).String(), cReset)
					go func() {
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
	for index, retryTopic := range retriesTopic {
		go func(topic, queue string) {
			_, err := eventBus.Conn.QueueSubscribe(topic, queue, func(msg *stan.Msg) {
				var data IntegrationEvent
				if err := json.Unmarshal(msg.Data, &data); err != nil {
					log.Printf("Error when Unmarshal event: %v", err)
				} else {
					for _, handler := range eventBus.Subscriptions[data.Type] {
						item := handler
						if reflect.TypeOf(handler.EventHandler).String() != data.HandleBy {
							continue
						}
						concrete := reflect.ValueOf(item.FallbackHandler)
						log.Printf("%s[Queue]%s Handle event Id [%s] redelivery [%d] from topic [%s%s%s] with FallbackHandler [%s%v%s]", cBlue, cReset, data.ID,
							data.RedeliveryCount, cCyan, topic, cReset, cCyan, reflect.TypeOf(item).String(), cReset)

						go func() {
							if data.IsRetry {
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
		}(retryTopic, fmt.Sprintf("queue-%d", index))
	}
}
