package EventBus

import (
	"encoding/json"
	"log"
	"reflect"
	"strconv"

	"github.com/nats-io/stan.go"
)

type IEventBus interface {
	Publish(event Event, topic string)
	Subscribe(event Event, eventHandlers []*EventHandler, topic string)
}

type EventBus struct {
	Subscriptions map[reflect.Type][]interface{}
	Conn          stan.Conn
	ClientId      string
}
type ResultError struct {
	Res interface{}
	Err error
}

func (eventBus *EventBus) New() *EventBus {
	nc, err := stan.Connect("local-cluster", eventBus.ClientId, stan.NatsURL("localhost:4222"))
	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}
	return &EventBus{Subscriptions: make(map[reflect.Type][]interface{}), Conn: nc}
}

func (eventBus *EventBus) PublishAsync(event interface{}, topic string) {
	data, err := json.Marshal(&event)
	if err != nil {
		log.Printf("error when marshal event %v", err)
	}
	if msgId, err := eventBus.Conn.PublishAsync(topic, data, func(str string, err error) {
		log.Println("what ", str)
	}); err != nil {
		log.Printf("error when publish event %v", err)
	} else {
		log.Printf("Delivery message [%s] success to topic [%s]", msgId, topic)
	}
}

func (eventBus *EventBus) Publish(event interface{}, topic string) {
	data, err := json.Marshal(&event)
	if err != nil {
		log.Printf("error when marshal event %v", err)
	}
	if err := eventBus.Conn.Publish(topic, data); err != nil {
		log.Printf("error when publish event %v", err)
	}
	log.Printf("Delivery message success to topic [%s]", topic)
}

func (eventBus *EventBus) QueueSubscribe(event interface{}, eventHandlers []interface{}, topic, queue string) {
	if _, ok := eventBus.Subscriptions[reflect.TypeOf(event)]; !ok {
		eventBus.Subscriptions[reflect.TypeOf(event)] = eventHandlers
	} else {
		eventBus.Subscriptions[reflect.TypeOf(event)] = append(eventBus.Subscriptions[reflect.TypeOf(event)], eventHandlers...)
	}

	go func() {
		_, err := eventBus.Conn.QueueSubscribe(topic, queue, func(msg *stan.Msg) {
			resultError := make(chan ResultError)
			for _, handler := range eventBus.Subscriptions[reflect.TypeOf(event)] {
				item := handler
				concrete := reflect.ValueOf(item)
				go func() {
					j, _ := json.Marshal(msg.MsgProto)
					log.Println(string(j))
					log.Printf("[Queue] Handle event seg [%s] from topic [%s] with handler [%v]", strconv.Itoa(int(msg.Sequence)), topic, reflect.TypeOf(item).String())
					concrete.MethodByName("Handle").Call([]reflect.Value{reflect.ValueOf(msg.Data), reflect.ValueOf(resultError)})
				}()
			}

			for range eventBus.Subscriptions[reflect.TypeOf(event)] {
				result := <-resultError
				if result.Err != nil {
					log.Printf("Error when handle topic [%s]. Error --> %v", topic, result.Err)
				}
			}
		}, stan.DurableName(queue))
		if err != nil {
			log.Printf("error when subscribe topic [%s]. Error --> %v", topic, err)
		}
	}()
}

func (eventBus *EventBus) Subscribe(event interface{}, eventHandlers []interface{}, topic string) {
	if _, ok := eventBus.Subscriptions[reflect.TypeOf(event)]; !ok {
		eventBus.Subscriptions[reflect.TypeOf(event)] = eventHandlers
	} else {
		eventBus.Subscriptions[reflect.TypeOf(event)] = append(eventBus.Subscriptions[reflect.TypeOf(event)], eventHandlers...)
	}

	go func() {
		_, err := eventBus.Conn.Subscribe(topic, func(msg *stan.Msg) {
			resultError := make(chan ResultError)
			for _, handler := range eventBus.Subscriptions[reflect.TypeOf(event)] {
				item := handler
				concrete := reflect.ValueOf(item)
				go func() {
					log.Printf("Handle event seg [%s] from topic [%s] with handler [%v]", strconv.Itoa(int(msg.Sequence)), topic, reflect.TypeOf(item).String())
					concrete.MethodByName("Handle").Call([]reflect.Value{reflect.ValueOf(msg.Data), reflect.ValueOf(resultError)})
				}()
			}

			for range eventBus.Subscriptions[reflect.TypeOf(event)] {
				result := <-resultError
				if result.Err != nil {
					log.Printf("Error when handle topic [%s]. Error --> %v", topic, result.Err)
				}
			}
		})
		if err != nil {
			log.Printf("error when subscribe topic [%s]. Error --> %v", topic, err)
		}
	}()
}
