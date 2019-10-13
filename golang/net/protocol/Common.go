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
var HandShakeSignature = []byte{11, 3, 72, 10, 4, 75, 6, 4, 04, 1, 9, 6, 12, 15, 10, 3, 1, 4, 1, 5, 9, 2, 6, 5}
