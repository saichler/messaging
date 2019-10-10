package mailbox

import . "github.com/saichler/utils/golang"

type SourceMultiPartMessages struct {
	multiPartMessages *ConcurrentMap
}

func newSourceMultiPartMessages() *SourceMultiPartMessages {
	smp := &SourceMultiPartMessages{}
	smp.multiPartMessages = NewConcurrentMap()
	return smp
}

func (smp *SourceMultiPartMessages) newMultiPartMessage(messageID uint32) *MultiPartMessage {
	mpm := &MultiPartMessage{}
	mpm.packets = NewList()
	smp.multiPartMessages.Put(messageID, mpm)
	return mpm
}

func (smp *SourceMultiPartMessages) getMultiPartMessage(messageID uint32) *MultiPartMessage {
	var mpm *MultiPartMessage
	exist, ok := smp.multiPartMessages.Get(messageID)
	if !ok {
		mpm = smp.newMultiPartMessage(messageID)
	} else {
		mpm = exist.(*MultiPartMessage)
	}
	return mpm
}

func (smp *SourceMultiPartMessages) deleteMultiPartMessage(messageID uint32) {
	smp.multiPartMessages.Del(messageID)
}
