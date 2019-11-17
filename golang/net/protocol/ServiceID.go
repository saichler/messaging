package protocol

import (
	"bytes"
	"fmt"
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

func (sid *ServiceID) Bytes(ba *ByteSlice) {
	sid.networkID.Bytes(ba)
	ba.AddString(sid.topic)
	ba.AddUInt16(sid.id)
}

func (sid *ServiceID) Object(ba *ByteSlice) {
	sid.networkID = &NetworkID{}
	sid.networkID.Object(ba)
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
	buff.WriteString("[T=")
	buff.WriteString(sid.topic)
	buff.WriteString(",D=")
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

func (serviceID *ServiceID) Parse(str string) error {
	serviceID.networkID = &NetworkID{}
	e := serviceID.networkID.Parse(str)
	if e != nil {
		fmt.Println(serviceID.networkID.String())
		return e
	}

	serviceID.topic = GetTagValue(str, "T")
	idString := GetTagValue(str, "D")
	id, e := strconv.Atoi(idString)
	if e != nil {
		return e
	}
	serviceID.id = uint16(id)
	return nil
}
