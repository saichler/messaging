package protocol

const (
	PUBLISH_MARK     = -9999
	UNREACHABLE_MARK = -9998
	SWITCH_PORT      = 52000
	MAX_PORT         = 54000
)

var PUBLISH_HID = newPublishHabitatID()
var UNREACH_HID = newDestUnreachableHabitatID()
var MTU = 512
var KEY = "bNhDNirkahDbiJJirSfaNNEXDprtwQoK"
var ENCRYPTED = true
var HANDSHAK_DATA = []byte{127, 83, 83, 127, 12, 10, 11}
