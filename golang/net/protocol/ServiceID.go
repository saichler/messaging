package protocol

import (
	"bytes"
	. "github.com/saichler/utils/golang"
	"strconv"
)

type ServiceID struct {
	networkID *NetworkID
	topic     string
	id        uint16
}

func NewServiceID(networkID *NetworkID, topic string, id uint16) *ServiceID {
	sid := &ServiceID{}
	sid.networkID = networkID
	sid.topic = topic
	sid.id = id
	return sid
}

func (sid *ServiceID) Marshal(ba *ByteSlice) {
	sid.networkID.Marshal(ba)
	ba.AddString(sid.topic)
	ba.AddUInt16(sid.id)
}

func (sid *ServiceID) Unmarshal(ba *ByteSlice) {
	sid.networkID = &NetworkID{}
	sid.networkID.Unmarshal(ba)
	sid.topic = ba.GetString()
	sid.id = ba.GetUInt16()
}

func (sid *ServiceID) Publish() bool {
	return sid.networkID.Publish()
}

func (sid *ServiceID) Unreachable() bool {
	return sid.networkID.Unreachable()
}

func (sid *ServiceID) String() string {
	buff := bytes.Buffer{}
	buff.WriteString(sid.networkID.String())
	buff.WriteString("[")
	buff.WriteString(sid.topic)
	buff.WriteString("]")
	buff.WriteString("[")
	buff.WriteString(strconv.Itoa(int(sid.id)))
	buff.WriteString("]")
	return buff.String()
}

func (sid *ServiceID) NetworkID() *NetworkID {
	return sid.networkID
}

func (sid *ServiceID) Topic() string {
	return sid.topic
}

func (sid *ServiceID) ID() uint16 {
	return sid.id
}
