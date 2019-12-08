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
	p := NewPacket(NewPacketHeader(source, NetConfig.UnreachableID(), false, false, 0), 0, 0, data)
	networkConnection.mailbox.PushOutbox(p.ToBytes(), priority)
}

func (networkSwitch *NetworkSwitch) handlePacket(data []byte, networkConnection *NetworkConnection) error {
	header := &PacketHeader{}
	bs := NewByteSliceWithData(data, 0)
	header.Read(bs)
	if header.Destination().Equal(NetConfig.UnreachableID()) {
		if header.Source().Equal(networkSwitch.netowrkNode.networkID) {
			networkSwitch.handleMyPacket(header, bs, networkConnection, true)
		}
	} else if header.Destination().Publish() {
		networkSwitch.handleMulticast(header, data, bs, networkConnection)
	} else if header.Destination().Equal(networkSwitch.netowrkNode.networkID) {
		networkSwitch.handleMyPacket(header, bs, networkConnection, false)
	} else {
		in := networkSwitch.getNetworkConnection(header.Destination())
		if in == nil {
			networkSwitch.sendUnreachable(header.Source(), header.Priority(), data)
			return errors.New("Unreachable address:" + header.Destination().String())
		}
		in.mailbox.PushOutbox(data, header.Priority())
	}
	return nil
}

func (networkSwitch *NetworkSwitch) handleMulticast(header *PacketHeader, data []byte, bs *ByteSlice, networkConnection *NetworkConnection) {
	if networkSwitch.netowrkNode.isNetworkSwitch {
		all := networkSwitch.getAllInternal()
		for k, v := range all {
			if !k.Equal(header.Source()) {
				v.mailbox.PushOutbox(data, header.Priority())
			}
		}
		if header.Source().Host() == networkSwitch.netowrkNode.networkID.Host() {
			all := networkSwitch.getAllExternal()
			for _, v := range all {
				v.mailbox.PushOutbox(data, header.Priority())
			}
		}
	}
	networkSwitch.handleMyPacket(header, bs, networkConnection, false)
}

func (networkSwitch *NetworkSwitch) handleMyPacket(header *PacketHeader, bs *ByteSlice, networkConnection *NetworkConnection, isUnreachable bool) {
	message := &Message{}
	p := &Packet{}
	p.SetHeader(header)
	p.Read(bs)
	networkConnection.DecodeMessage(p, message, isUnreachable)

	if message.Complete() {
		ne := networkSwitch.getNetworkConnection(header.Source())
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
