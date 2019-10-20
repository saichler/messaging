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

func (networkConnection *NetworkConnection) write(data []byte) error {
	start := time.Now().UnixNano()
	dataSize := len(data)
	size := Size(dataSize)
	data = append(size[0:], data...)
	dataSize = len(data)

	networkConnection.statistics.AddTxPackets(data)

	n, e := networkConnection.connection.Write(data)

	end := time.Now().UnixNano()

	networkConnection.statistics.AddTxTime(end - start)

	if e != nil || n != dataSize {
		msg := "Failed to send data: " + e.Error()
		Error(msg)
		return errors.New(msg)
	}
	return nil
}

func (networkConnection *NetworkConnection) read(size int) ([]byte, error) {
	data := make([]byte, size)
	n, e := networkConnection.connection.Read(data)

	if !networkConnection.networkNode.running {
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
		left, e := networkConnection.read(size - n)
		if e != nil {
			return nil, Error("Failed to read packet size", e)
		}
		data = append(data, left...)
	}

	return data, nil
}

func (networkConnection *NetworkConnection) nextPacket() error {
	pSize, e := networkConnection.read(4)
	if pSize == nil || e != nil {
		return e
	}

	size := int(binary.LittleEndian.Uint32(pSize))

	data, e := networkConnection.read(size)
	if data == nil || e != nil {
		return e
	}

	if networkConnection.networkNode.running {
		networkConnection.mailbox.PushInbox(data, Priority(data))
	}
	return nil
}

func (networkConnection *NetworkConnection) addPacketToOutbox(p *Packet) (error) {
	start := time.Now().UnixNano()
	data := p.Marshal()
	end := time.Now().UnixNano()
	networkConnection.statistics.AddTxTimeSync(end - start)
	networkConnection.mailbox.PushOutbox(data, p.Priority())
	return nil
}

func (networkConnection *NetworkConnection) newInterfacePacket(destination *ServiceID, messageID, packetID uint32, multi, persistence bool, priority int, data []byte) *Packet {
	if destination != nil {
		return NewPacket(networkConnection.networkNode.networkID, destination.NetworkID(), messageID, packetID, multi, persistence, priority, data)
	}
	return NewPacket(networkConnection.networkNode.networkID, nil, messageID, packetID, multi, persistence, priority, data)
}

func (networkConnection *NetworkConnection) runIncomming() {
	for ; networkConnection.networkNode.running; {
		err := networkConnection.nextPacket()
		if err != nil {
			Error("Error reading from socket:", err)
			break
		}
	}
	Info("Read Interface from:" + networkConnection.peerNetworkID.String() + " was shutdown!")
	Info("Statistics:")
	Info(networkConnection.statistics.String())
	networkConnection.isClosed = true
}

func (networkConnection *NetworkConnection) runOutgoing() {
	for ; networkConnection.networkNode.running; {
		data := networkConnection.mailbox.PopOutbox()
		err := networkConnection.write(data)
		if err != nil {
			Error("Error Sending to socket:", err)
			break
		}
	}
	Info("Write Interface to:" + networkConnection.peerNetworkID.String() + " was shutdown!")
	networkConnection.isClosed = true
}

func (networkConnection *NetworkConnection) deserializeMux() {
	time.Sleep(time.Second)
	for ; networkConnection.networkNode.running; {
		data := networkConnection.mailbox.PopInbox()
		if data != nil {
			networkConnection.statistics.AddRxPackets(data)
			networkConnection.networkNode.networkSwitch.handlePacket(data, networkConnection)
		} else {
			break
		}
	}
	Info("Handle Interface of:" + networkConnection.peerNetworkID.String() + " was shutdown!")
}

func (networkConnection *NetworkConnection) start() {
	go networkConnection.runIncomming()
	go networkConnection.runOutgoing()
	for i := 0; i < NetConfig.Deserialize(); i++ {
		go networkConnection.deserializeMux()
	}
}

func (networkConnection *NetworkConnection) handshake() (bool, error) {
	Info("Starting handshake process for:" + networkConnection.networkNode.networkID.String())

	packet := networkConnection.newInterfacePacket(nil, 0, 0, false, false, 0, NetConfig.Handshake())

	sendData := packet.Marshal()
	networkConnection.write(sendData)

	err := networkConnection.nextPacket()
	if err != nil {
		return false, err
	}

	data := networkConnection.mailbox.PopInbox()

	source, destination, multi, persist, priority, ba := UnmarshalHeaderOnly(data)
	p := &Packet{}
	p.Unmarshal(source, destination, multi, persist, priority, ba)

	Info("handshaked "+networkConnection.networkNode.networkID.String()+" with nid:", p.Source().String())
	networkConnection.peerNetworkID = p.Source()
	if networkConnection.peerNetworkID.Host() != networkConnection.networkNode.networkID.Host() {
		networkConnection.external = true
	}

	if networkConnection.peerNetworkID.Port() == NetConfig.SwitchPort() {
		networkConnection.networkNode.switchNetworkID = networkConnection.peerNetworkID
	}

	networkConnection.mailbox.SetName(networkConnection.peerNetworkID.String())

	added := networkConnection.networkNode.networkSwitch.addNetworkConnection(networkConnection)

	return added, nil
}

func (networkConnection *NetworkConnection) Shutdown() {
	networkConnection.connection.Close()
	networkConnection.mailbox.Shutdown()
}
