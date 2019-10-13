package netnode

import (
	"errors"
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"net"
	"strconv"
)

func bind() (net.Listener, int, error) {
	port := SWITCH_PORT

	Debug("Trying to bind to switch port " + strconv.Itoa(port) + ".");
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(port))

	if e != nil {
		for ; port < MAX_PORT && e != nil; port++ {
			Debug("Trying to bind to port " + strconv.Itoa(port) + ".")
			s, err := net.Listen("tcp", ":"+strconv.Itoa(port))
			e = err
			socket = s
			if e == nil {
				break
			}
		}
		Debug("Successfuly binded to port " + strconv.Itoa(port))
	}

	if port >= MAX_PORT {
		return nil, -1, errors.New("Failed to find an available port to bind to")
	}

	return socket, port, nil
}

func Size(s int) [4]byte {
	size := [4]byte{}
	size[0] = byte(s)
	size[1] = byte(s >> 8)
	size[2] = byte(s >> 16)
	size[3] = byte(s >> 24)
	return size
}
