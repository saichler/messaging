package netnode

import (
	"encoding/binary"
	"errors"
	. "github.com/saichler/messaging/golang/net/mailbox"
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"net"
	"strconv"
	"time"
)

type NetworkConnection struct {
	networkNode   *NetworkNode
	peerNetworkID *NetworkID
	connection    net.Conn
	external      bool
	mailbox       *Mailbox
	statistics    *NetworkInterfaceStatistics
	isClosed      bool
}

func newNetworkConnection(connection net.Conn, networkNode *NetworkNode) *NetworkConnection {
	networkConnection := &NetworkConnection{}
	networkConnection.connection = connection
	networkConnection.networkNode = networkNode
	networkConnection.mailbox = NewMailbox()
	networkConnection.statistics = newNetworkInterfaceStatistics()
	return networkConnection
}

func (in *NetworkConnection) write(data []byte) error {
	start := time.Now().UnixNano()
	dataSize := len(data)
	size := Size(dataSize)
	data = append(size[0:], data...)
	dataSize = len(data)

	in.statistics.AddTxPackets(data)

	n, e := in.connection.Write(data)

	end := time.Now().UnixNano()

	in.statistics.AddTxTime(end - start)

	if e != nil || n != dataSize {
		msg := "Failed to send data: " + e.Error()
		Error(msg)
		return errors.New(msg)
	}
	return nil
}

func (in *NetworkConnection) read(size int) ([]byte, error) {
	data := make([]byte, size)
	n, e := in.connection.Read(data)

	if !in.networkNode.running {
		return nil, nil
	}

	if e != nil {
		return nil, Error("Failed to read date size", e)
	}

	if n < size {
		if n == 0 {
			Warn("Expected " + strconv.Itoa(size) + " bytes but only read 0, Sleeping a second...")
			time.Sleep(time.Second)
		}
		data = data[0:n]
		left, e := in.read(size - n)
		if e != nil {
			return nil, Error("Failed to read packet size", e)
		}
		data = append(data, left...)
	}

	return data, nil
}

func (in *NetworkConnection) nextPacket() error {
	pSize, e := in.read(4)
	if pSize == nil || e != nil {
		return e
	}

	size := int(binary.LittleEndian.Uint32(pSize))

	data, e := in.read(size)
	if data == nil || e != nil {
		return e
	}

	if in.networkNode.running {
		in.mailbox.PushInbox(data, Priority(data))
	}
	return nil
}

func (in *NetworkConnection) addPacketToOutbox(p *Packet) (error) {
	start := time.Now().UnixNano()
	data := p.Marshal()
	end := time.Now().UnixNano()
	in.statistics.AddTxTimeSync(end - start)
	in.mailbox.PushOutbox(data, p.Priority())
	return nil
}

func (in *NetworkConnection) newInterfacePacket(destination *ServiceID, messageID, packetID uint32, multi, persistence bool, priority int, data []byte) *Packet {
	if destination != nil {
		return NewPacket(in.networkNode.networkID, destination.NetworkID(), messageID, packetID, multi, persistence, priority, data)
	}
	return NewPacket(in.networkNode.networkID, nil, messageID, packetID, multi, persistence, priority, data)
}

func (in *NetworkConnection) runIncomming() {
	for ; in.networkNode.running; {
		err := in.nextPacket()
		if err != nil {
			Error("Error reading from socket:", err)
			break
		}
	}
	Info("Read Interface from:" + in.peerNetworkID.String() + " was shutdown!")
	Info("Statistics:")
	Info(in.statistics.String())
	in.isClosed = true
}

func (in *NetworkConnection) runOutgoing() {
	for ; in.networkNode.running; {
		data := in.mailbox.PopOutbox()
		err := in.write(data)
		if err != nil {
			Error("Error Sending to socket:", err)
			break
		}
	}
	Info("Write Interface to:" + in.peerNetworkID.String() + " was shutdown!")
	in.isClosed = true
}

func (in *NetworkConnection) handle() {
	time.Sleep(time.Second)
	for ; in.networkNode.running; {
		data := in.mailbox.PopInbox()
		if data != nil {
			in.statistics.AddRxPackets(data)
			in.networkNode.nSwitch.handlePacket(data, in.mailbox)
		} else {
			break
		}
	}
	Info("Handle Interface of:" + in.peerNetworkID.String() + " was shutdown!")
}

func (in *NetworkConnection) start() {
	go in.runIncomming()
	go in.runOutgoing()
	go in.handle()
}

func (in *NetworkConnection) handshake() (bool, error) {
	Info("Starting handshake process for:" + in.networkNode.networkID.String())

	packet := in.newInterfacePacket(nil, 0, 0, false, false, 0, HandShakeSignature)

	sendData := packet.Marshal()
	in.write(sendData)

	err := in.nextPacket()
	if err != nil {
		return false, err
	}

	data := in.mailbox.PopInbox()

	source, destination, multi, persist, priority, ba := UnmarshalHeaderOnly(data)
	p := &Packet{}
	p.Unmarshal(source, destination, multi, persist, priority, ba)

	Info("handshaked "+in.networkNode.networkID.String()+" with nid:", p.Source().String())
	in.peerNetworkID = p.Source()
	if in.peerNetworkID.Host() != in.networkNode.networkID.Host() {
		in.external = true
	}

	if in.peerNetworkID.Port() == SWITCH_PORT {
		in.networkNode.switchNetworkID = in.peerNetworkID
	}

	in.mailbox.SetName(in.peerNetworkID.String())

	added := in.networkNode.nSwitch.addInterface(in)

	return added, nil
}

func (in *NetworkConnection) Shutdown() {
	in.connection.Close()
	in.mailbox.Shutdown()
}
