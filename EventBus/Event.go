package EventBus

type Event struct {
	ID      string
	Payload interface{}
}

type EventHandler struct {
}
