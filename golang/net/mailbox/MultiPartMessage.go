package mailbox

import . "github.com/saichler/utils/golang"

type MultiPartMessage struct {
	messageID            uint32
	packets              *List
	totalExpectedPackets uint32
	byteLength           uint32
}
