package netnode

import (
	"errors"
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
)

type NetworkSwitch struct {
	netowrkNode *NetworkNode
	internal    *ConcurrentMap
	external    *ConcurrentMap
}

func newSwitch(networkNode *NetworkNode) *NetworkSwitch {
	nSwitch := &NetworkSwitch{}
	nSwitch.internal = NewConcurrentMap()
	nSwitch.external = NewConcurrentMap()
	nSwitch.netowrkNode = networkNode
	return nSwitch
}

func (networkSwitch *NetworkSwitch) removeNetowrkConnection(networkConection *NetworkConnection) {
	if !networkConection.external {
		networkSwitch.internal.Del(*networkConection.peerNetworkID)
	} else {
		networkSwitch.external.Del(networkConection.peerNetworkID.Host())
	}
	Info("Network Connection " + networkConection.peerNetworkID.String() + " was deleted")
}

func (networkSwitch *NetworkSwitch) addNetworkConnection(networkConnection *NetworkConnection) bool {
	if !networkConnection.external {
		old, _ := networkSwitch.internal.Get(*networkConnection.peerNetworkID)
		if old != nil {
			networkSwitch.removeNetowrkConnection(old.(*NetworkConnection))
		}
		networkSwitch.internal.Put(*networkConnection.peerNetworkID, networkConnection)
	} else {
		old, _ := networkSwitch.external.Get(networkConnection.peerNetworkID.Host())
		if old != nil {
			networkSwitch.removeNetowrkConnection(old.(*NetworkConnection))
		}
		networkSwitch.external.Put(networkConnection.peerNetworkID.Host(), networkConnection)
	}
	return true
}

func (networkSwitch *NetworkSwitch) getNetworkConnection(networkID *NetworkID) *NetworkConnection {
	var networkConnection *NetworkConnection
	if networkID.Host() == networkSwitch.netowrkNode.networkID.Host() {
		if networkSwitch.netowrkNode.isNetworkSwitch {
			inter, _ := networkSwitch.internal.Get(*networkID)
			if inter == nil {
				return nil
			}
			networkConnection = inter.(*NetworkConnection)
		} else {
			inter, _ := networkSwitch.internal.Get(*networkSwitch.netowrkNode.switchNetworkID)
			if inter == nil {
				return nil
			}
			networkConnection = inter.(*NetworkConnection)
		}
	} else {
		if networkSwitch.netowrkNode.isNetworkSwitch {
			inter, _ := networkSwitch.external.Get(networkID.Host())
			if inter == nil {
				return nil
			}
			networkConnection = inter.(*NetworkConnection)
		} else {
			inter, _ := networkSwitch.internal.Get(*networkSwitch.netowrkNode.switchNetworkID)
			if inter == nil {
				return nil
			}
			networkConnection = inter.(*NetworkConnection)
		}
	}
	return networkConnection
}

func (networkSwitch *NetworkSwitch) sendUnreachable(source *NetworkID, priority int, data []byte) {
	networkConnection := networkSwitch.getNetworkConnection(source)
	if networkConnection == nil {
		return
	}
	p := NewPacket(source, NetConfig.UnreachableID(), 0, 0, false, false, 0, data)
	networkConnection.mailbox.PushOutbox(p.Bytes(), priority)
}

func (networkSwitch *NetworkSwitch) handlePacket(data []byte, networkConnection *NetworkConnection) error {
	source, destination, multi, persist, priority, ba := Header(data)
	if destination.Equal(NetConfig.UnreachableID()) {
		if source.Equal(networkSwitch.netowrkNode.networkID) {
			networkSwitch.handleMyPacket(source, destination, multi, persist, priority, data, ba, networkConnection, true)

		}
	} else if destination.Publish() {
		networkSwitch.handleMulticast(source, destination, multi, persist, priority, data, ba, networkConnection)
	} else if destination.Equal(networkSwitch.netowrkNode.networkID) {
		networkSwitch.handleMyPacket(source, destination, multi, persist, priority, data, ba, networkConnection, false)
	} else {
		in := networkSwitch.getNetworkConnection(destination)
		if in == nil {
			networkSwitch.sendUnreachable(source, priority, data)
			return errors.New("Unreachable address:" + destination.String())
		}
		in.mailbox.PushOutbox(data, priority)
	}
	return nil
}

func (networkSwitch *NetworkSwitch) handleMulticast(source, destination *NetworkID, multi, persist bool, priority int, data []byte, ba *ByteSlice, networkConnection *NetworkConnection) {
	if networkSwitch.netowrkNode.isNetworkSwitch {
		all := networkSwitch.getAllInternal()
		for k, v := range all {
			if !k.Equal(source) {
				v.mailbox.PushOutbox(data, priority)
			}
		}
		if source.Host() == networkSwitch.netowrkNode.networkID.Host() {
			all := networkSwitch.getAllExternal()
			for _, v := range all {
				v.mailbox.PushOutbox(data, priority)
			}
		}
	}
	networkSwitch.handleMyPacket(source, destination, multi, persist, priority, data, ba, networkConnection, false)
}

func (networkSwitch *NetworkSwitch) handleMyPacket(source, destination *NetworkID, multi, persist bool, priority int, data []byte, ba *ByteSlice, networkConnection *NetworkConnection, isUnreachable bool) {
	message := &Message{}
	p := &Packet{}
	p.Object(source, destination, multi, persist, priority, ba)
	networkConnection.DecodeMessage(p, message, isUnreachable)

	if message.Complete() {
		ne := networkSwitch.getNetworkConnection(source)
		ne.statistics.AddRxMessages()
		if !isUnreachable {
			networkSwitch.netowrkNode.messageHandler.HandleMessage(message)
		} else {
			networkSwitch.netowrkNode.messageHandler.HandleUnreachable(message)
		}
	}
}

func (networkSwitch *NetworkSwitch) getAllInternal() map[NetworkID]*NetworkConnection {
	result := make(map[NetworkID]*NetworkConnection)
	all := networkSwitch.internal.GetMap()
	for k, v := range all {
		key := k.(NetworkID)
		value := v.(*NetworkConnection)
		if !value.isClosed {
			result[key] = value
		}
	}
	return result
}

func (networkSwitch *NetworkSwitch) getAllExternal() map[int32]*NetworkConnection {
	result := make(map[int32]*NetworkConnection)
	all := networkSwitch.external.GetMap()
	for k, v := range all {
		key := k.(int32)
		value := v.(*NetworkConnection)
		if !value.isClosed {
			result[key] = value
		}
	}
	return result
}

func (networkSwitch *NetworkSwitch) shutdown() {
	allInternal := networkSwitch.getAllInternal()
	for _, v := range allInternal {
		v.Shutdown()
	}
	allExternal := networkSwitch.getAllExternal()
	for _, v := range allExternal {
		v.Shutdown()
	}
}

func (networkSwitch *NetworkSwitch) multicastFromSwitch(message *Message) {
	faulty := make([]*NetworkConnection, 0)
	internal := networkSwitch.getAllInternal()
	for _, in := range internal {
		err := in.SendMessage(message)
		if err != nil {
			faulty = append(faulty, in)
		}
	}
	if message.Source().NetworkID().Host() == networkSwitch.netowrkNode.networkID.Host() {
		external := networkSwitch.getAllExternal()
		for _, in := range external {
			err := in.SendMessage(message)
			if err != nil {
				faulty = append(faulty, in)
			}
		}
	}
	for _, in := range faulty {
		networkSwitch.removeNetowrkConnection(in)
	}
}
