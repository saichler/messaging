package mailbox

import (
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
)

type Mailbox struct {
	pending *ConcurrentMap
	inbox   *PriorityQueue
	outbox  *PriorityQueue
}

func NewMailbox() *Mailbox {
	mb := &Mailbox{}
	mb.inbox = NewPriorityQueue()
	mb.outbox = NewPriorityQueue()
	mb.pending = NewConcurrentMap()
	return mb
}

func (mailbox *Mailbox) SetName(name string) {
	mailbox.inbox.SetName(name)
	mailbox.outbox.SetName(name)
}

func (mailbox *Mailbox) PopInbox() []byte {
	next := mailbox.inbox.Pop()
	if next != nil {
		return next.([]byte)
	}
	return nil
}

func (mailbox *Mailbox) PopOutbox() []byte {
	next := mailbox.outbox.Pop()
	if next != nil {
		return next.([]byte)
	}
	return nil
}

func (mailbox *Mailbox) PushInbox(pData []byte, priority int) {
	mailbox.inbox.Push(pData, priority)
}

func (mailbox *Mailbox) PushOutbox(pData []byte, priority int) {
	mailbox.outbox.Push(pData, priority)
}

func (mailbox *Mailbox) getMultiPartMessage(packet *Packet) (*MultiPartMessage, *SourceMultiPartMessages) {
	sourceNID := packet.Source()
	var smpm *SourceMultiPartMessages
	existing, ok := mailbox.pending.Get(*sourceNID)
	if !ok {
		smpm = newSourceMultiPartMessages()
		mailbox.pending.Put(*sourceNID, smpm)
	} else {
		smpm = existing.(*SourceMultiPartMessages)
	}
	mpm := smpm.getMultiPartMessage(packet.MessageID())
	return mpm, smpm
}

func (mailbox *Mailbox) AddPacket(packet *Packet) ([]byte, bool) {
	mp, smp := mailbox.getMultiPartMessage(packet)
	mp.packets.Add(packet)
	if mp.totalExpectedPackets == 0 && packet.PacketID() == 0 {
		ba := NewByteSliceWithData(packet.Data(), 0)
		mp.totalExpectedPackets = ba.GetUInt32()
		mp.byteLength = ba.GetUInt32()
	}

	isComplete := false
	if mp.totalExpectedPackets > 0 && mp.packets.Size() == int(mp.totalExpectedPackets) {
		isComplete = true
	}

	if isComplete {
		messageData := make([]byte, int(mp.byteLength))
		for i := 0; i < int(mp.totalExpectedPackets); i++ {
			qp := mp.packets.Get(i).(*Packet)
			if qp.PacketID() != 0 {
				start := int((qp.PacketID() - 1) * uint32(MTU))
				end := start + len(qp.Data())
				copy(messageData[start:end], qp.Data()[:])
			}
		}
		smp.deleteMultiPartMessage(packet.MessageID())
		return messageData, isComplete
	}
	return nil, isComplete
}

func (mailbox *Mailbox) Shutdown() {
	mailbox.inbox.Shutdown()
	mailbox.outbox.Shutdown()
}
