package protocol

import . "github.com/saichler/utils/golang"

type PacketHeader struct {
	source      *NetworkID
	destination *NetworkID
	multi       bool
	persisted   bool
	priority    int //uint6 if existed
}

func NewPacketHeader(source, destination *NetworkID, multi, persisted bool, priority int) *PacketHeader {
	packetHeader := &PacketHeader{}
	packetHeader.source = source
	packetHeader.destination = destination
	packetHeader.multi = multi
	packetHeader.persisted = persisted
	packetHeader.priority = priority
	return packetHeader
}

func (packetHeader *PacketHeader) Source() *NetworkID {
	return packetHeader.source
}

func (packetHeader *PacketHeader) Destination() *NetworkID {
	return packetHeader.destination
}

func (packetHeader *PacketHeader) Multi() bool {
	return packetHeader.multi
}

func (packetHeader *PacketHeader) Persisted() bool {
	return packetHeader.persisted
}

func (packetHeader *PacketHeader) Priority() int {
	return packetHeader.priority
}

func (packetHeader *PacketHeader) ToBytes() []byte {
	bs := NewByteSlice()
	packetHeader.Write(bs)
	return bs.Data()
}

func (packetHeader *PacketHeader) FromBytes(data []byte) {
	bs := NewByteSliceWithData(data, 0)
	packetHeader.Read(bs)
}

func (packetHeader *PacketHeader) Write(bs *ByteSlice) {
	packetHeader.source.Write(bs)
	packetHeader.destination.Write(bs)
	mpp := Encode2BoolAndUInt6(packetHeader.multi, packetHeader.persisted, packetHeader.priority)
	bs.AddByte(mpp)
}

func (packetHeader *PacketHeader) Read(bs *ByteSlice) {
	packetHeader.source = &NetworkID{}
	packetHeader.destination = &NetworkID{}
	packetHeader.source.Read(bs)
	packetHeader.destination.Read(bs)
	mpp := bs.GetByte()
	packetHeader.multi, packetHeader.persisted, packetHeader.priority = Decode2BoolAndUInt6(mpp)
}
