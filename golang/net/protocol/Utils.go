package protocol

import (
	. "github.com/saichler/security"
	"log"
	"net"
	"strconv"
	"strings"
)
import . "github.com/saichler/utils/golang"

func encrypt(data []byte) []byte {
	if NetConfig.Encrypted() {
		encData, err := Encode(data, NetConfig.EncKey())
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
	if NetConfig.Encrypted() {
		decryData, err := Decode(data, NetConfig.EncKey())
		if err != nil {
			panic("Failed to decrypt data!")
			return data
		} else {
			return decryData
		}
	}
	return data
}

func NewLocalNetworkID(port int32) *NetworkID {
	return NewNetworkID(GetLocalIpAddress(), port)
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

func GetIpAsInt32(ipaddr string) (int32, error) {
	var ipint int32
	arr := strings.Split(ipaddr, ".")
	ipint = 0
	a, e := strconv.Atoi(arr[0])
	if e != nil {
		return -1, e
	}
	b, e := strconv.Atoi(arr[1])
	if e != nil {
		return -1, e
	}
	c, e := strconv.Atoi(arr[2])
	if e != nil {
		return -1, e
	}
	d, e := strconv.Atoi(strings.Split(arr[3], "/")[0])
	if e != nil {
		return -1, e
	}
	ipint += int32(a) << 24
	ipint += int32(b) << 16
	ipint += int32(c) << 8
	ipint += int32(d)
	return ipint, nil
}

func Priority(data []byte) int {
	p := 0
	_, _, p = Decode2BoolAndUInt6(data[32])
	return p
}

func NewPublishServiceID(topic string) *ServiceID {
	return NewServiceID(NetConfig.publishID, topic, 0)
}
