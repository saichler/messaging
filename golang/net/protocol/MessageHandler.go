package protocol

type MessageHandler interface {
	HandleMessage(*Message)
	HandleUnreachable(*Message)
}
