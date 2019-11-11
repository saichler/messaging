package netnode

import (
	"fmt"
	"github.com/saichler/messaging/golang/net/protocol"
	utils "github.com/saichler/utils/golang"
	"net"
	"time"
)

func (netNode *NetworkNode) listenForUDPBroadcast() {
	utils.Info("Starting UDP listener on 40299")

	bs := utils.NewByteSlice()
	netNode.networkID.Marshal(bs)
	data := bs.Data()
	broadcast, err := net.ResolveUDPAddr("udp", "255.255.255.255:40299")

	addr, err := net.ResolveUDPAddr("udp", ":40299")
	if err != nil {
		return
	}

	go netNode.waitForBroadcast(addr, len(data))

	conn, err := net.DialUDP("udp4", nil, broadcast)
	if err != nil {
		return
	}

	for {
		_, err := conn.Write(data)
		if err != nil {
			break
		}
		time.Sleep(time.Second * 5)
	}
}

func (netNode *NetworkNode) waitForBroadcast(addr *net.UDPAddr, size int) {
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	for {
		data := make([]byte, size)
		n, err := conn.Read(data)
		if err != nil {
			return
		}
		if n == size {
			go netNode.receive(data)
		}
	}
}

func (netNode *NetworkNode) receive(data []byte) {
	bs := utils.NewByteSliceWithData(data, 0)
	nid := &protocol.NetworkID{}
	nid.Unmarshal(bs)
	if !nid.Equal(netNode.networkID) {
		fmt.Println("Receive ping from:", nid.String())
	}
}
