package netnode

import (
	"errors"
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"net"
	"strconv"
	"sync"
	"time"
)

type NetworkNode struct {
	networkID       *NetworkID
	messageHandler  MessageHandler
	isNetworkSwitch bool
	networkSwitch   *NetworkSwitch
	socket          net.Listener
	port            int32
	switchNetworkID *NetworkID
	lock            *sync.Cond
	nextMessageID   uint32
	active          bool
}

func NewNetworkNode(handler MessageHandler) (*NetworkNode, error) {
	networkNode := &NetworkNode{}
	networkNode.lock = sync.NewCond(&sync.Mutex{})
	networkNode.active = true
	networkNode.messageHandler = handler

	socket, port, e := bind()

	if e != nil {
		Error("Failed to create a network node:", e)
		return nil, e
	} else {
		networkNode.networkID = NewLocalNetworkID(port)
		Debug("Bounded to ", networkNode.networkID.String())
		networkNode.isNetworkSwitch = port == NetConfig.SwitchPort()
		if !networkNode.isNetworkSwitch {
			e := networkNode.uplinkToSwitch()
			for ; e != nil; {
				time.Sleep(time.Second * 5)
				e = networkNode.uplinkToSwitch()
			}
		}
	}
	networkNode.socket = socket
	networkNode.port = port
	networkNode.networkSwitch = newSwitch(networkNode)
	networkNode.start()
	if networkNode.isNetworkSwitch {
		go networkNode.listenForUDPBroadcast()
	}
	return networkNode, nil
}

func (networkNode *NetworkNode) Shutdown() {
	networkNode.active = false
	net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(int(networkNode.networkID.Port())))
	networkNode.networkSwitch.shutdown()
	networkNode.lock.L.Lock()
	networkNode.lock.Broadcast()
	networkNode.lock.L.Unlock()
}

func (networkNode *NetworkNode) start() {
	go networkNode.waitForlinks()
	time.Sleep(time.Second / 5)
}

func (networkNode *NetworkNode) waitForlinks() {
	if networkNode.active {
		Info("Habitat ", networkNode.networkID.String(), " is waiting for links")
	}
	for ; networkNode.active; {
		connection, error := networkNode.socket.Accept()
		if !networkNode.active {
			break
		}
		if error != nil {
			Fatal("Failed to accept a new connection from socket: ", error)
			return
		}
		//add a new interface
		go networkNode.newNetworkConnection(connection)
	}
	networkNode.socket.Close()
	Info("Habitat:" + networkNode.networkID.String() + " was shutdown!")
}

func (networkNode *NetworkNode) newNetworkConnection(c net.Conn) (*NetworkID, error) {
	Debug("connecting to: " + c.RemoteAddr().String())
	networkConnection := newNetworkConnection(c, networkNode)
	added, e := networkConnection.handshake()
	if e != nil {
		Error("Failed to add network connection:", e)
	}

	if e != nil || !added {
		return nil, e
	}

	networkConnection.start()

	return networkConnection.peerNetworkID, nil
}

func (networkNode *NetworkNode) uplinkToSwitch() error {
	switchPortString := strconv.Itoa(int(NetConfig.SwitchPort()))
	c, e := net.Dial("tcp", "127.0.0.1:"+switchPortString)
	if e != nil {
		Error("Failed to open connection to switch: ", e)
		return e
	}
	go networkNode.newNetworkConnection(c)
	return e
}

func (networkNode *NetworkNode) Uplink(host string) *NetworkID {
	switchPortString := strconv.Itoa(int(NetConfig.SwitchPort()))
	c, e := net.Dial("tcp", host+":"+switchPortString)
	if e != nil {
		Error("Failed to open connection to host: "+host, e)
	}
	networkID, err := networkNode.newNetworkConnection(c)
	if err != nil {
		return nil
	}
	return networkID
}

func (networkNode *NetworkNode) waitForUplinkToSwitch() *NetworkConnection {
	networkNode.lock.L.Lock()
	defer networkNode.lock.L.Unlock()
	networkConnection := networkNode.networkSwitch.getNetworkConnection(networkNode.switchNetworkID)
	if networkConnection == nil || networkConnection.isClosed {
		Error("Uplink to switch is closed, trying to open a new one.")
		e := networkNode.uplinkToSwitch()
		for ; e != nil; {
			time.Sleep(time.Second * 5)
			e = networkNode.uplinkToSwitch()
		}
	}
	time.Sleep(time.Second)
	networkConnection = networkNode.networkSwitch.getNetworkConnection(networkNode.switchNetworkID)
	return networkConnection
}

func (networkNode *NetworkNode) SendMessage(message *Message) error {
	var e error
	if message.Destination().NetworkID().Equal(networkNode.networkID) {
		networkNode.messageHandler.HandleMessage(message)
	} else if message.Publish() {
		if !message.Source().NetworkID().Equal(networkNode.networkID) {
			return errors.New("Multicast Message Cannot be forward!")
		}
		networkNode.messageHandler.HandleMessage(message)
		if networkNode.isNetworkSwitch {
			networkNode.networkSwitch.multicastFromSwitch(message)
		} else {
			networkConnection := networkNode.networkSwitch.getNetworkConnection(networkNode.switchNetworkID)
			if networkConnection == nil || networkConnection.isClosed {
				networkConnection = networkNode.waitForUplinkToSwitch()
			}
			e = networkConnection.SendMessage(message)
			if e != nil {
				Error("Failed to send multicast message:", e)
			}
		}
	} else {
		networkConnection := networkNode.networkSwitch.getNetworkConnection(message.Destination().NetworkID())
		if networkConnection == nil {
			Error("Unknown Destination:" + message.Destination().String())
			networkNode.messageHandler.HandleUnreachable(message)
			return errors.New("Unknown Destination:" + message.Destination().String())
		}
		e = networkConnection.SendMessage(message)
		if e != nil {
			Error("Failed to send message:", e)
		}
	}
	return e
}

func (networkNode *NetworkNode) SwitchNetworkID() *NetworkID {
	return networkNode.switchNetworkID
}

func (networkNode *NetworkNode) NetworkID() *NetworkID {
	return networkNode.networkID
}

func (networkNode *NetworkNode) ServiceID() *ServiceID {
	return NewServiceID(networkNode.networkID, "", 0)
}

func (networkNode *NetworkNode) NextMessageID() uint32 {
	networkNode.lock.L.Lock()
	defer networkNode.lock.L.Unlock()
	result := networkNode.nextMessageID
	networkNode.nextMessageID++
	return result
}

func (networkNode *NetworkNode) NewMessage(source, destination, origin *ServiceID, topic string, priority int, data []byte, isReply bool) *Message {
	return NewMessage(source, destination, origin, networkNode.NextMessageID(), topic, priority, data, isReply)
}

func (networkNode *NetworkNode) WaitForShutdown() {
	networkNode.lock.L.Lock()
	networkNode.lock.Wait()
	networkNode.lock.L.Unlock()
}

func (networkNode *NetworkNode) Port() int32 {
	return networkNode.port
}
