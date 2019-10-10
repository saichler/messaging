package node

import (
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"net"
	"sync"
)

type NetworkNode struct {
	networkID             *NetworkID
	messageHandler        MessageHandler
	isNetworkSwitch       bool
	networkSwitch         *NetworkSwitch
	socket                net.Listener
	machineSwitchNetowkID *NetworkID
	lock                  *sync.Cond
	nextMessageID         uint32
	running               bool
}

func NewNetworkNode(handler MessageHandler) (*NetworkNode, error) {
	nn := &NetworkNode{}
	nn.lock = sync.NewCond(&sync.Mutex{})
	nn.running = true
	nn.messageHandler = handler

	socket, port, e := bind()

	if e != nil {
		Error("Failed to create a network node:", e)
		return nil, e
	} else {
		nn.networkID = NewLocalNetworkID(port)
		Debug("Bounded to ", nn.networkID.String())
		nn.isNetworkSwitch = port == SWITCH_PORT
		if !nn.isNetworkSwitch {
			/*
				e := habitat.uplinkToSwitch()
				for ; e != nil; {
					time.Sleep(time.Second * 5)
					e = habitat.uplinkToSwitch()
				}*/
		}
	}
	nn.socket = socket
	nn.networkSwitch = newSwitch(nn)
	//habitat.start()
	return habitat, nil
}
