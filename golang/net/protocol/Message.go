package protocol

import (
	. "github.com/saichler/utils/golang"
	"strconv"
)

type Message struct {
	source         *ServiceID
	destination    *ServiceID
	originalSource *ServiceID
	messageID      uint32
	topic          string
	data           []byte
	priority       int
	complete       bool
	unreachable    bool
}

func (message *Message) Marshal() []byte {
	ba := NewByteSlice()
	message.source.Marshal(ba)
	message.destination.Marshal(ba)
	message.originalSource.Marshal(ba)
	ba.AddUInt32(message.messageID)
	ba.AddString(message.topic)
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
	message.data = ba.GetByteSlice()
}

func (message *Message) Publish() bool {
	return message.destination.Publish()
}

func (message *Message) Decode(pkt *Packet, inbox *Mailbox, isUnreachable bool) {

	packet := pkt

	if isUnreachable {
		origSource, origDest, om, oprs, opri, ba := unmarshalPacketHeader(pkt.Data)
		packet.UnmarshalAll(origSource, origDest, om, oprs, opri, ba)
	}

	if packet.MultiPart {
		message.Data, message.Complete = inbox.addPacket(packet)
	} else {
		message.Data = packet.Data
		message.Complete = true
	}

	if message.Complete {
		if packet.Dest.Equal(UNREACH_HID) {
		} else {
			message.Unmarshal(packet.Source, packet.Dest)
		}
	}
}

func (message *Message) Send(ne *Interface) error {
	ne.statistics.AddTxMessages()

	messageData := message.Marshal()

	if len(messageData) > MTU {

		totalParts := len(messageData) / MTU
		left := len(messageData) - totalParts*MTU

		if left > 0 {
			totalParts++
		}

		totalParts++

		if totalParts > 1000 {
			Info("Large Message, total parts:" + strconv.Itoa(totalParts))
		}

		ba := ByteSlice{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(messageData)))

		packet := ne.CreatePacket(message.Dest, message.MID, 0, true, 0, ba.Data())
		err := ne.sendPacket(packet)
		if err != nil {
			return err
		}

		for i := 0; i < totalParts-1; i++ {
			loc := i * MTU
			var packetData []byte
			if i < totalParts-2 || left == 0 {
				packetData = messageData[loc : loc+MTU]
			} else {
				packetData = messageData[loc : loc+left]
			}

			packet := ne.CreatePacket(message.Dest, message.MID, uint32(i+1), true, 0, packetData)
			if i%1000 == 0 {
				Info("Sent " + strconv.Itoa(i) + " packets out of " + strconv.Itoa(totalParts))
			}
			err = ne.sendPacket(packet)
			if err != nil {
				Error("Was able to send only" + strconv.Itoa(i) + " packets")
				break
			}
		}
	} else {
		packet := ne.CreatePacket(message.Dest, message.MID, 0, false, 0, messageData)
		ne.sendPacket(packet)
	}

	return nil
}
