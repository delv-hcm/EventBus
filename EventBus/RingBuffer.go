package EventBus

import "github.com/nats-io/stan.go"

type RingBuffer struct {
	InChan  chan ResultError
	OutChan chan ResultError
}

type CallbackFn func(result *ResultError, msg *stan.Msg)

func (ring *RingBuffer) New(inChan chan ResultError, outChan chan ResultError) *RingBuffer {
	return &RingBuffer{InChan: inChan, OutChan: outChan}
}

func (ring *RingBuffer) Run(callback CallbackFn) {
	for v := range ring.InChan {
		select {
		case ring.OutChan <- v:
			res := <-ring.OutChan
			callback(&res, v.Msg)
		default:
			<-ring.OutChan
			ring.OutChan <- v
			res := <-ring.OutChan
			callback(&res, v.Msg)
		}
	}
	close(ring.OutChan)
}
