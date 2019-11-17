package protocol

import (
	. "github.com/saichler/utils/golang"
)

type Packet struct {
	source      *NetworkID
	destination *NetworkID
	multi       bool
	persisted   bool
	priority    int //uint6 if existed
	messageID   uint32
	packetID    uint32
	data        []byte
}

func NewPacket(source, destination *NetworkID, messageID, packetID uint32, multi, persisted bool, priority int, data []byte) *Packet {
	packet := &Packet{}
	packet.source = source
	packet.destination = destination
	packet.messageID = messageID
	packet.packetID = packetID
	packet.multi = multi
	packet.persisted = persisted
	packet.priority = priority
	packet.data = data
	return packet
}

func (p *Packet) Source() *NetworkID {
	return p.source
}

func (p *Packet) Destination() *NetworkID {
	return p.destination
}

func (p *Packet) Multi() bool {
	return p.multi
}

func (p *Packet) Persisted() bool {
	return p.persisted
}

func (p *Packet) Priority() int {
	return p.priority
}

func (p *Packet) MessageID() uint32 {
	return p.messageID
}

func (p *Packet) PacketID() uint32 {
	return p.packetID
}

func (p *Packet) Data() []byte {
	return p.data
}

func (p *Packet) Bytes() []byte {
	bs := NewByteSlice()
	p.source.Bytes(bs)
	p.destination.Bytes(bs)
	mpp := Encode2BoolAndUInt6(p.multi, p.persisted, p.priority)
	bs.AddByte(mpp)
	bs.AddUInt32(p.messageID)
	bs.AddUInt32(p.packetID)
	bs.AddByteSlice(encrypt(p.data))
	return bs.Data()
}

func Header(data []byte) (*NetworkID, *NetworkID, bool, bool, int, *ByteSlice) {
	ba := NewByteSliceWithData(data, 0)
	source := &NetworkID{}
	dest := &NetworkID{}
	source.Object(ba)
	dest.Object(ba)
	mpp := ba.GetByte()
	m, prs, pri := Decode2BoolAndUInt6(mpp)
	return source, dest, m, prs, pri, ba
}

func (p *Packet) Object(source, dest *NetworkID, m, prs bool, pri int, ba *ByteSlice) {
	p.source = source
	p.destination = dest
	p.multi = m
	p.persisted = prs
	p.priority = pri
	p.messageID = ba.GetUInt32()
	p.packetID = ba.GetUInt32()
	p.data = decrypt(ba.GetByteSlice())
}
