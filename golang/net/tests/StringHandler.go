package tests

import (
	. "github.com/saichler/messaging/golang/net/netnode"
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"sync"
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
	if !message.IsReply() {
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
		dest = NewServiceID(smh.node.SwitchNetworkID(), "", 0)
	}
	source := NewServiceID(smh.node.NetworkID(), "", 0)
	message := smh.node.NewMessage(source, dest, source, "StringTest", 0, []byte(str), false)
	smh.node.SendMessage(message)
}

func (sfh *StringMessageHandler) ReplyString(str string, node *NetworkNode, dest *ServiceID) {
	if sfh.print {
		Debug("Sending Reply:" + str + " to:" + dest.String())
	}
	if dest == nil {
		dest = NewServiceID(node.SwitchNetworkID(), "", 0)
	}
	source := NewServiceID(node.NetworkID(), "", 0)
	message := node.NewMessage(source, dest, source, "StringTest", 0, []byte(str), true)

	node.SendMessage(message)
}
