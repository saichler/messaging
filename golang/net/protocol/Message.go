package protocol

import (
	. "github.com/saichler/utils/golang"
)

type Message struct {
	source         *ServiceID
	destination    *ServiceID
	originalSource *ServiceID
	messageID      uint32
	topic          string
	data           []byte
	priority       int
	isReply        bool

	complete    bool
	unreachable bool
}

func NewMessage(source, destination, originalSource *ServiceID, messageID uint32, topic string, priority int, data []byte, isReply bool) *Message {
	message := &Message{}
	message.source = source
	message.destination = destination
	message.originalSource = originalSource
	message.messageID = messageID
	message.topic = topic
	message.priority = priority
	message.data = data
	message.isReply = isReply
	return message
}

func (message *Message) Marshal() []byte {
	ba := NewByteSlice()
	message.source.Marshal(ba)
	message.destination.Marshal(ba)
	message.originalSource.Marshal(ba)
	ba.AddUInt32(message.messageID)
	ba.AddString(message.topic)
	ba.AddBool(message.isReply)
	ba.AddByteSlice(message.data)
	return ba.Data()
}

func (message *Message) Unmarshal(ba *ByteSlice) {
	message.source = &ServiceID{}
	message.destination = &ServiceID{}
	message.originalSource = &ServiceID{}
	message.source.Unmarshal(ba)
	message.destination.Unmarshal(ba)
	message.originalSource.Unmarshal(ba)
	message.messageID = ba.GetUInt32()
	message.topic = ba.GetString()
	message.isReply = ba.GetBool()
	message.data = ba.GetByteSlice()
}

func (message *Message) Source() *ServiceID {
	return message.source
}

func (message *Message) Destination() *ServiceID {
	return message.destination
}

func (message *Message) OriginalSource() *ServiceID {
	return message.originalSource
}

func (message *Message) MessageID() uint32 {
	return message.messageID
}

func (message *Message) Topic() string {
	return message.topic
}

func (message *Message) Data() []byte {
	return message.data
}

func (message *Message) SetData(data []byte) {
	message.data = data
}

func (message *Message) Complete() bool {
	return message.complete
}

func (message *Message) SetComplete(complete bool) {
	message.complete = complete
}

func (message *Message) Priority() int {
	return message.priority
}

func (message *Message) Unreachable() bool {
	return message.unreachable
}

func (message *Message) Publish() bool {
	return message.destination.Publish()
}

func (message *Message) IsReply() bool {
	return message.isReply
}
