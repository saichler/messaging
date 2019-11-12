package netnode

import (
	"github.com/saichler/messaging/golang/net/protocol"
	"github.com/saichler/security"
	. "github.com/saichler/utils/golang"
	"net"
	"time"
)

type UdpBroadcastListener struct {
	pings   *Map
	node    *NetworkNode
	running bool
}

func NewUDPListener(node *NetworkNode) *UdpBroadcastListener {
	ubl := &UdpBroadcastListener{}
	ubl.node = node
	ubl.pings = NewMap()
	return ubl
}

func (ubl *UdpBroadcastListener) start() error {
	Info("Starting UDP Broadcast listener on 40299")

	ubl.running = true
	bs := NewByteSlice()
	ubl.node.NetworkID().Marshal(bs)
	data := bs.Data()
	encData, e := security.Encode(data, protocol.NetConfig.EncKey())
	if e != nil {
		return Error("Failed to encode broadcast packet")
	}

	broadcast, err := net.ResolveUDPAddr("udp", "255.255.255.255:40299")

	addr, err := net.ResolveUDPAddr("udp", ":40299")
	if err != nil {
		return err
	}

	go ubl.waitForBroadcasts(addr)

	conn, err := net.DialUDP("udp4", nil, broadcast)
	if err != nil {
		return Error("Failed to create broadcast client:", e)
	}

	for ubl.running {
		_, err := conn.Write(encData)
		if err != nil {
			return Error("Failed to write broadcast packet:", e)
		}
		for i := 0; i < 10 && ubl.running; i++ {
			time.Sleep(time.Second / 2)
		}
	}
	Info("UDP Broadcast stopped")
	return nil
}

func (ubl *UdpBroadcastListener) waitForBroadcasts(addr *net.UDPAddr) {
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	for ubl.running {
		encData := make([]byte, 256)
		n, err := conn.Read(encData)
		if err != nil {
			return
		}

		encData = encData[0:n]
		data, e := security.Decode(encData, protocol.NetConfig.EncKey())
		if e != nil {
			Error("UDP Broadcast Failed to decode data")
		} else {
			ubl.receive(data)
		}
	}
}

func (ubl *UdpBroadcastListener) receive(data []byte) {
	bs := NewByteSliceWithData(data, 0)
	nid := &protocol.NetworkID{}
	nid.Unmarshal(bs)
	if nid.Equal(ubl.node.NetworkID()) {
		return
	}
	if nid.Host() > ubl.node.NetworkID().Host() {
		ubl.checkForUplink(nid)
		return
	}
	count, ok := ubl.pings.Get(nid.Host()).(int)
	if !ok {
		ubl.pings.Put(nid.Host(), 1)
		return
	}
	count++
	ubl.pings.Put(nid.Host(), count)
	if count >= 2 {
		ubl.checkForUplink(nid)
	}
}

func (ubl *UdpBroadcastListener) checkForUplink(nid *protocol.NetworkID) {
	nc := ubl.node.networkSwitch.getNetworkConnection(nid)
	if nc == nil {
		ip := protocol.GetIpAsString(nid.Host())
		Info("UDP Broadcast creating an uplink to: " + ip)
		ubl.node.Uplink(ip)
	}
}
