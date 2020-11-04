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
	Rb            *RingBuffer
}
type ResultError struct {
	Res interface{}
	Err error
	Seq uint64
	Msg *stan.Msg
}

func (eventBus *EventBus) New(size int) *EventBus {
	nc, err := stan.Connect("local-cluster", eventBus.ClientId, stan.NatsURL("localhost:4222"))
	if err != nil {
		log.Fatalf("Can't connect: %v\n", err)
	}
	return &EventBus{Subscriptions: make(map[reflect.Type][]interface{}), Conn: nc,
		Rb: &RingBuffer{InChan: make(chan ResultError), OutChan: make(chan ResultError, size)}}
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

// QueueSubscribe
func (eventBus *EventBus) QueueSubscribe(event interface{}, eventHandlers []interface{}, topic, queue string) {
	if _, ok := eventBus.Subscriptions[reflect.TypeOf(event)]; !ok {
		eventBus.Subscriptions[reflect.TypeOf(event)] = eventHandlers
	} else {
		eventBus.Subscriptions[reflect.TypeOf(event)] = append(eventBus.Subscriptions[reflect.TypeOf(event)], eventHandlers...)
	}

	go eventBus.Rb.Run(func(result *ResultError, rawMsg *stan.Msg) {
		if result.Err != nil {
			log.Println(result.Err)
			// should publish to retry topic
		}
	})

	go func() {
		_, err := eventBus.Conn.QueueSubscribe(topic, queue, func(msg *stan.Msg) {
			for _, handler := range eventBus.Subscriptions[reflect.TypeOf(event)] {
				item := handler
				concrete := reflect.ValueOf(item)
				go func() {
					log.Printf("[Queue] Handle event seg [%s] redeliveryCount [%d] from topic [%s] with handler [%v]", strconv.Itoa(int(msg.Sequence)),
						msg.RedeliveryCount, topic, reflect.TypeOf(item).String())
					concrete.MethodByName("Handle").Call([]reflect.Value{reflect.ValueOf(msg), reflect.ValueOf(eventBus.Rb.InChan)})
				}()
			}
		}, stan.DurableName(queue))
		if err != nil {
			log.Printf("error when subscribe topic [%s]. Error --> %v", topic, err)
		}
	}()
}
