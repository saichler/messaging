package protocol

import (
	. "github.com/saichler/security"
	"log"
	"net"
	"strconv"
	"strings"
)
import . "github.com/saichler/utils/golang"

func encrypt(data []byte) ([]byte) {
	if ENCRYPTED {
		encData, err := Encode(data, KEY)
		if err != nil {
			Error("Failed to encrypt data, sending unsecure!", err)
			return data
		} else {
			return encData
		}
	}
	return data
}

func decrypt(data []byte) []byte {
	if ENCRYPTED {
		decryData, err := Decode(data, KEY)
		if err != nil {
			panic("Failed to decrypt data!")
			return data
		} else {
			return decryData
		}
	}
	return data
}

func NewLocalNetworkID(port int) *NetworkID {
	return NewNetworkID(GetLocalIpAddress(), port)
}

func newPublishHabitatID() *NetworkID {
	newHID := &NetworkID{}
	newHID.most = PUBLISH_MARK
	newHID.less = PUBLISH_MARK
	return newHID
}

func newDestUnreachableHabitatID() *NetworkID {
	newHID := &NetworkID{}
	newHID.most = UNREACHABLE_MARK
	newHID.less = UNREACHABLE_MARK
	return newHID
}

func GetLocalIpAddress() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal("Unable to access interfaces\n", err)
	}
	for _, _interface := range ifaces {
		intName := strings.ToLower(_interface.Name)
		if !strings.Contains(intName, "lo") &&
			!strings.Contains(intName, "br") &&
			!strings.Contains(intName, "vir") {
			intAddresses, err := _interface.Addrs()
			if err != nil {
				log.Fatal("Unable to access interface address\n", err)
			}

			for _, address := range intAddresses {
				return address.String()
			}
		}
	}
	return ""
}

func GetIpAsString(ip int32) string {
	a := strconv.FormatInt(int64((ip>>24)&0xff), 10)
	b := strconv.FormatInt(int64((ip>>16)&0xff), 10)
	c := strconv.FormatInt(int64((ip>>8)&0xff), 10)
	d := strconv.FormatInt(int64(ip&0xff), 10)
	return a + "." + b + "." + c + "." + d
}

func GetIpAsInt32(ipaddr string) int32 {
	var ipint int32
	arr := strings.Split(ipaddr, ".")
	ipint = 0
	a, _ := strconv.Atoi(arr[0])
	b, _ := strconv.Atoi(arr[1])
	c, _ := strconv.Atoi(arr[2])
	d, _ := strconv.Atoi(strings.Split(arr[3], "/")[0])
	ipint += int32(a) << 24
	ipint += int32(b) << 16
	ipint += int32(c) << 8
	ipint += int32(d)
	return ipint
}

func Priority(data []byte) int {
	p := 0
	_, _, p = Decode2BoolAndUInt6(data[32])
	return p
}
