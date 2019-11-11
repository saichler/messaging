package netnode

import (
	"fmt"
	"net"
	"time"
)

func ListenForUDPBroadcast() {
	addr, err := net.ResolveUDPAddr("udp", ":40299")
	if err != nil {
		return
	}
	go waitForBroadcast(addr)
	time.Sleep(time.Second * 5)
	data := make([]byte, 16)

	broadcast, err := net.ResolveUDPAddr("udp", "255.255.255.255:40299")

	conn, err := net.DialUDP("udp4", nil, broadcast)
	for {
		time.Sleep(time.Second * 5)
		fmt.Println("sending")
		_, err := conn.Write(data)
		if err != nil {
			break
		}
	}
}

func waitForBroadcast(addr *net.UDPAddr) {
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	for {
		data := make([]byte, 16)
		_, err = conn.Read(data)
		if err != nil {
			return
		}
		go receive(conn)
	}
}

func receive(conn *net.UDPConn) {
	fmt.Println("received local:", conn.LocalAddr())
	fmt.Println("received remote:", conn.LocalAddr())
}
