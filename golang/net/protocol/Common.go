package protocol

type NetworkNodeConfig struct {
	publishId     int64
	publishID     *NetworkID
	unreachableId int64
	unreachableID *NetworkID
	switchPort    int32
	maxSwitchPort int32
	mtu           int
	encrypted     bool
	encKey        string
	handShake     []byte
	deserialize   int
}

func newNetConfig() *NetworkNodeConfig {
	netconfig := &NetworkNodeConfig{}

	netconfig.publishId = -9999
	netconfig.publishID = &NetworkID{}
	netconfig.publishID.most = netconfig.publishId
	netconfig.publishID.less = netconfig.publishId

	netconfig.unreachableId = -9998
	netconfig.unreachableID = &NetworkID{}
	netconfig.unreachableID.most = netconfig.unreachableId
	netconfig.unreachableID.less = netconfig.unreachableId

	netconfig.switchPort = 52000
	netconfig.maxSwitchPort = 54000
	netconfig.mtu = 512
	netconfig.encrypted = true
	netconfig.encKey = "bNhDNirkahDbiJJirSfaNNEXDprtwQoK"
	netconfig.handShake = []byte{11, 3, 72, 10, 4, 75, 6, 4, 04, 1, 9, 6, 12, 15, 10, 3, 1, 4, 1, 5, 9, 2, 6, 5}
	netconfig.deserialize = 1

	return netconfig
}

func (ncfg *NetworkNodeConfig) SwitchPort() int32 {
	return ncfg.switchPort
}

func (ncfg *NetworkNodeConfig) MaxSwitchPort() int32 {
	return ncfg.maxSwitchPort
}

func (ncfg *NetworkNodeConfig) MTU() int {
	return ncfg.mtu
}

func (ncfg *NetworkNodeConfig) Encrypted() bool {
	return ncfg.encrypted
}

func (ncfg *NetworkNodeConfig) EncKey() string {
	return ncfg.encKey
}

func (ncfg *NetworkNodeConfig) Handshake() []byte {
	return ncfg.handShake
}

func (ncfg *NetworkNodeConfig) Deserialize() int {
	return ncfg.deserialize
}

func (ncfg *NetworkNodeConfig) PublishId() int64 {
	return ncfg.publishId
}

func (ncfg *NetworkNodeConfig) PublishID() *NetworkID {
	return ncfg.publishID
}

func (ncfg *NetworkNodeConfig) UnreachableId() int64 {
	return ncfg.unreachableId
}

func (ncfg *NetworkNodeConfig) UnreachableID() *NetworkID {
	return ncfg.unreachableID
}

func (ncfg *NetworkNodeConfig) Set(mtu int, encrypted bool) {
	ncfg.mtu = mtu
	ncfg.encrypted = encrypted
}

var NetConfig = newNetConfig()

/*
const (
	PUBLISH_MARK     = -9999
	UNREACHABLE_MARK = -9998
	SWITCH_PORT      = 52000
	MAX_PORT         = 54000
)

var PublishNetworkID = newPublishHabitatID()
var UnreachableNetworkID = newDestUnreachableHabitatID()
var MTU = 512
var KEY = "bNhDNirkahDbiJJirSfaNNEXDprtwQoK"
var ENCRYPTED = true
var HandShakeSignature = []byte{11, 3, 72, 10, 4, 75, 6, 4, 04, 1, 9, 6, 12, 15, 10, 3, 1, 4, 1, 5, 9, 2, 6, 5}
var NumberOfDeserializeMux = 1
*/
