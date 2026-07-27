package driven

type WebSocketMessage struct {
	Type      string
	Channel   string
	Payload   interface{}
	Timestamp int64
}

type WebSocketPort interface {
	Broadcast(raw []byte)
	SendDirect(userID string, raw []byte)
	PublishChannel(channel string, raw []byte)
	Run()
	StartRedisListener()
}
