package netnode

import (
	"errors"
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"net"
	"strconv"
)

func bind() (net.Listener, int32, error) {
	port := NetConfig.SwitchPort()

	Debug("Trying to bind to switch port ", strconv.Itoa(int(port)), ".");
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(int(port)))

	if e != nil {
		for ; port < NetConfig.MaxSwitchPort() && e != nil; port++ {
			Debug("Trying to bind to port " + strconv.Itoa(int(port)) + ".")
			s, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
			e = err
			socket = s
			if e == nil {
				break
			}
		}
		Debug("Successfuly binded to port " + strconv.Itoa(int(port)))
	}

	if port >= NetConfig.MaxSwitchPort() {
		return nil, -1, errors.New("Failed to find an available port to bind to")
	}

	return socket, port, nil
}

func Size(s int) [4]byte {
	size := [4]byte{}
	size[0] = byte(s)
	size[1] = byte(s >> 8)
	size[2] = byte(s >> 16)
	size[3] = byte(s >> 24)
	return size
}

func (networkConnection *NetworkConnection) DecodeMessage(p *Packet, m *Message, isUnreachable bool) {

	var messageData []byte
	var messageComplete bool

	if isUnreachable {
		origSource, origDest, multi, persist, priority, ba := UnmarshalHeaderOnly(p.Data())
		p.Unmarshal(origSource, origDest, multi, persist, priority, ba)
	}

	if p.Multi() {
		messageData, messageComplete = networkConnection.mailbox.AddPacket(p)
	} else {
		messageData = p.Data()
		messageComplete = true
	}

	m.SetComplete(messageComplete)

	if messageComplete {
		if p.Destination().Equal(NetConfig.UnreachableID()) {
		} else {
			ba := NewByteSliceWithData(messageData, 0)
			m.Unmarshal(ba)
		}
	}
}

func (networkConnection *NetworkConnection) SendMessage(message *Message) error {

	networkConnection.statistics.AddTxMessages()
	messageData := message.Marshal()

	mtu := NetConfig.MTU()

	if len(messageData) > mtu {

		totalParts := len(messageData) / mtu
		left := len(messageData) - totalParts*mtu

		if left > 0 {
			totalParts++
		}

		totalParts++

		ba := ByteSlice{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(messageData)))

		packet := networkConnection.newInterfacePacket(message.Destination(), message.MessageID(), 0, true, false, 0, ba.Data())
		err := networkConnection.addPacketToOutbox(packet)
		if err != nil {
			return err
		}

		for i := 0; i < totalParts-1; i++ {
			loc := i * mtu
			var packetData []byte
			if i < totalParts-2 || left == 0 {
				packetData = messageData[loc : loc+mtu]
			} else {
				packetData = messageData[loc : loc+left]
			}

			packet := networkConnection.newInterfacePacket(message.Destination(), message.MessageID(), uint32(i+1), true, false, 0, packetData)

			err = networkConnection.addPacketToOutbox(packet)
			if err != nil {
				Error("Was able to send only" + strconv.Itoa(i) + " packets")
				return err
			}
		}
	} else {
		packet := networkConnection.newInterfacePacket(message.Destination(), message.MessageID(), 0, false, false, 0, messageData)
		return networkConnection.addPacketToOutbox(packet)
	}

	return nil
}
