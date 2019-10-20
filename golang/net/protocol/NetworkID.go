package protocol

import (
	"bytes"
	. "github.com/saichler/utils/golang"
	"strconv"
)

type NetworkID struct {
	most int64
	less int64
}

func NewNetworkID(ipv4 string, port int32) *NetworkID {
	newHID := &NetworkID{}
	ip := GetIpAsInt32(ipv4)
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

func (networkID *NetworkID) Marshal(ba *ByteSlice) {
	if networkID != nil {
		ba.AddInt64(networkID.most)
		ba.AddInt64(networkID.less)
	} else {
		ba.AddInt64(0)
		ba.AddInt64(0)
	}
}

func (networkID *NetworkID) Unmarshal(ba *ByteSlice) {
	networkID.most = ba.GetInt64()
	networkID.less = ba.GetInt64()
}

func (networkID *NetworkID) String() string {
	ip := int32(networkID.most >> 32)
	port := int(networkID.less - ((networkID.less >> 32) << 32))
	buff := bytes.Buffer{}
	buff.WriteString("[UuidM=")
	buff.WriteString(strconv.Itoa(int(networkID.most)))
	buff.WriteString(",IP=")
	buff.WriteString(GetIpAsString(ip))
	buff.WriteString(",Port=")
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
