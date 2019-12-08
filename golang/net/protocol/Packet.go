package protocol

import (
	. "github.com/saichler/utils/golang"
)

type Packet struct {
	header    *PacketHeader
	messageID uint32
	packetID  uint32
	data      []byte
}

func NewPacket(header *PacketHeader, messageID, packetID uint32, data []byte) *Packet {
	packet := &Packet{}
	packet.header = header
	packet.messageID = messageID
	packet.packetID = packetID
	packet.data = data
	return packet
}

func (p *Packet) Header() *PacketHeader {
	return p.header
}

func (p *Packet) SetHeader(h *PacketHeader) {
	p.header = h
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

func (p *Packet) ToBytes() []byte {
	bs := NewByteSlice()
	p.Write(bs)
	return bs.Data()
}

func (p *Packet) FromBytes(data []byte) {
	bs := NewByteSliceWithData(data, 0)
	p.Read(bs)
}

func (p *Packet) Write(bs *ByteSlice) {
	p.header.Write(bs)
	bs.AddUInt32(p.messageID)
	bs.AddUInt32(p.packetID)
	bs.AddByteSlice(encrypt(p.data))
}

func (p *Packet) Read(bs *ByteSlice) {
	p.messageID = bs.GetUInt32()
	p.packetID = bs.GetUInt32()
	p.data = decrypt(bs.GetByteSlice())
}
