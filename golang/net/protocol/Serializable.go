package protocol

import . "github.com/saichler/utils/golang"

type Serializable interface {
	Unmarshal(ba *ByteSlice)
	Marshal(ba *ByteSlice)
}
