package protocol

import (
	"bytes"
	. "github.com/saichler/utils/golang"
	"strconv"
	"strings"
)

type NetworkID struct {
	most int64
	less int64
}

func NewNetworkID(ipv4 string, port int32) *NetworkID {
	newHID := &NetworkID{}
	ip, e := GetIpAsInt32(ipv4)
	if e != nil {
		panic(e)
	}
	newHID.most = 0;
	newHID.less = int64(ip)<<32 + int64(port)
	return newHID
}

func (networkID *NetworkID) Most() int64 {
	return networkID.most
}

func (networkID *NetworkID) Less() int64 {
	return networkID.less
}

func (networkID *NetworkID) Bytes(ba *ByteSlice) {
	if networkID != nil {
		ba.AddInt64(networkID.most)
		ba.AddInt64(networkID.less)
	} else {
		ba.AddInt64(0)
		ba.AddInt64(0)
	}
}

func (networkID *NetworkID) Object(ba *ByteSlice) {
	networkID.most = ba.GetInt64()
	networkID.less = ba.GetInt64()
}

func (networkID *NetworkID) String() string {
	ip := int32(networkID.less >> 32)
	port := int(networkID.less - ((networkID.less >> 32) << 32))
	buff := bytes.Buffer{}
	buff.WriteString("[M=")
	buff.WriteString(strconv.Itoa(int(networkID.most)))
	buff.WriteString(",Ip=")
	buff.WriteString(GetIpAsString(ip))
	buff.WriteString(",P=")
	buff.WriteString(strconv.Itoa(port))
	buff.WriteString("]")
	return buff.String()
}

func (networkID *NetworkID) Equal(other *NetworkID) bool {
	return networkID.most == other.most &&
		networkID.less == other.less
}

func (networkID *NetworkID) Publish() bool {
	if networkID.most == NetConfig.publishId {
		return true
	}
	return false
}

func (networkID *NetworkID) Unreachable() bool {
	if networkID.most == NetConfig.unreachableId {
		return true
	}
	return false
}

func (networkID *NetworkID) Host() int32 {
	return int32(networkID.less >> 32)
}

func (networkID *NetworkID) Port() int32 {
	return int32(networkID.less - ((networkID.less >> 32) << 32))
}

func (networkID *NetworkID) Parse(str string) error {
	most := GetTagValue(str, "M")
	ipString := GetTagValue(str, "Ip")
	portString := GetTagValue(str, "P")

	m, e := strconv.Atoi(most)
	if e != nil {
		return e
	}
	networkID.most = int64(m)

	ip, e := GetIpAsInt32(ipString)
	if e != nil {
		return e
	}
	port, e := strconv.Atoi(portString)
	if e != nil {
		return e
	}
	networkID.less = int64(ip)<<32 + int64(port)
	return nil
}

func GetTagValue(str, tag string) string {
	index := strings.Index(str, tag)
	if index != -1 {
		subst := str[index:]
		index = strings.Index(subst, "=")
		if index != -1 {
			subst = subst[index+1:]
			index1 := strings.Index(subst, ",")
			index2 := strings.Index(subst, "]")
			if index1 != -1 && index1 < index2 {
				return subst[0:index1]
			} else if index2 != -1 {
				return subst[0:index2]
			}
		}
	}
	return ""
}
