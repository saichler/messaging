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

func NewNetworkID(ipv4 string, port int) *NetworkID {
	newHID := &NetworkID{}
	ip := GetIpAsInt32(ipv4)
	newHID.most = 0;
	newHID.less = int64(ip)<<32 + int64(port)
	return newHID
}

func (nid *NetworkID) Most() int64 {
	return nid.most
}

func (nid *NetworkID) Less() int64 {
	return nid.less
}

func (nid *NetworkID) Marshal(ba *ByteSlice) {
	if nid != nil {
		ba.AddInt64(nid.most)
		ba.AddInt64(nid.less)
	} else {
		ba.AddInt64(0)
		ba.AddInt64(0)
	}
}

func (nid *NetworkID) Unmarshal(ba *ByteSlice) {
	nid.most = ba.GetInt64()
	nid.less = ba.GetInt64()
}

func (nid *NetworkID) String() string {
	ip := int32(nid.most >> 32)
	port := int(nid.less - ((nid.less >> 32) << 32))
	buff := bytes.Buffer{}
	buff.WriteString("[UuidM=")
	buff.WriteString(strconv.Itoa(int(nid.most)))
	buff.WriteString(",IP=")
	buff.WriteString(GetIpAsString(ip))
	buff.WriteString(",Port=")
	buff.WriteString(strconv.Itoa(port))
	buff.WriteString("]")
	return buff.String()
}

func (nid *NetworkID) Equal(other *NetworkID) bool {
	return nid.most == other.most &&
		nid.less == other.less
}

func (nid *NetworkID) Publish() bool {
	if nid.most == PUBLISH_MARK {
		return true
	}
	return false
}

func (nid *NetworkID) Unreachable() bool {
	if nid.most == UNREACHABLE_MARK {
		return true
	}
	return false
}
