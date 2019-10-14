package tests

import (
	. "github.com/saichler/messaging/golang/net/netnode"
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"sync"
)

const (
	REQUEST = "Request"
	REPLY   = "Reply"
)

var replyCount = 0
var unreachCount = 0
var myx = &sync.Mutex{}

type StringMessageHandler struct {
	node *NetworkNode

	print bool
}

func NewStringMessageHandler() *StringMessageHandler {
	sfh := &StringMessageHandler{}
	sfh.print = true
	node, e := NewNetworkNode(sfh)
	if e != nil {
		panic(e)
	}
	sfh.node = node
	return sfh
}

func (sfh *StringMessageHandler) HandleUnreachable(message *Message) {
	unreachCount++
	Info("Handled Unreachable!!!!")
}

func (sfh *StringMessageHandler) HandleMessage(message *Message) {
	str := string(message.Data())
	if message.Topic() == REQUEST {
		if sfh.print {
			Info("Request: " + str + " from:" + message.Source().String())
		}
		sfh.ReplyString(str, sfh.node, message.Source())
	} else {
		myx.Lock()
		replyCount++
		myx.Unlock()
		if sfh.print {
			Info("Reply: " + str + " to:" + message.Destination().String())
		}
	}
}

func (smh *StringMessageHandler) SendString(str string, dest *ServiceID) {
	if smh.print {
		Debug("Sending Request:" + str)
	}
	if dest == nil {
		dest = NewServiceID(smh.node.SwitchNetworkID(), "")
	}
	source := NewServiceID(smh.node.NetworkID(), "")
	message := smh.node.NewMessage(source, dest, source, REQUEST, 0, []byte(str))
	smh.node.SendMessage(message)
}

func (sfh *StringMessageHandler) ReplyString(str string, node *NetworkNode, dest *ServiceID) {
	if sfh.print {
		Debug("Sending Reply:" + str + " to:" + dest.String())
	}
	if dest == nil {
		dest = NewServiceID(node.SwitchNetworkID(), "")
	}
	source := NewServiceID(node.NetworkID(), "")
	message := node.NewMessage(source, dest, source, REPLY, 0, []byte(str))

	node.SendMessage(message)
}
