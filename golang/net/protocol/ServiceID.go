package protocol

import (
	"bytes"
	. "github.com/saichler/utils/golang"
)

type ServiceID struct {
	networkID   *NetworkID
	serviceName string
}

func NewServiceID(networkID *NetworkID, serviceName string) *ServiceID {
	sid := &ServiceID{}
	sid.networkID = networkID
	sid.serviceName = serviceName
	return sid
}

func (sid *ServiceID) Marshal(ba *ByteSlice) {
	sid.networkID.Marshal(ba)
	ba.AddString(sid.serviceName)
}

func (sid *ServiceID) Unmarshal(ba *ByteSlice) {
	sid.networkID = &NetworkID{}
	sid.networkID.Unmarshal(ba)
	sid.serviceName = ba.GetString()
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
	buff.WriteString("ServiceName=")
	buff.WriteString(sid.serviceName)
	buff.WriteString("]")
	return buff.String()
}

func (sid *ServiceID) NetworkID() *NetworkID {
	return sid.networkID
}

func (sid *ServiceID) ServiceName() string {
	return sid.serviceName
}
