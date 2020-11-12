package EventBus

import (
	"github.com/segmentio/ksuid"
)

type IntegrationEvent struct {
	ID              string
	Payload         []byte
	Delay           int
	RedeliveryCount int
	OriginSub       string
	Sub             string
	IsRetry         bool
	Type            string
}

func (e *IntegrationEvent) New() *IntegrationEvent {
	return &IntegrationEvent{ID: ksuid.New().String(), IsRetry: false, RedeliveryCount: 0, Delay: 0}
}

type IntegrationEventHandler struct {
}
