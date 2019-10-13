package netnode

import (
	"bytes"
	"strconv"
	"sync"
)

type NetworkInterfaceStatistics struct {
	txMessages int64
	rxMessages int64
	txPackets  int64
	rxPackets  int64
	txBytes    int64
	rxBytes    int64
	txTime     int64
	rxTime     int64
	mtx        *sync.Mutex
}

func newNetworkInterfaceStatistics() *NetworkInterfaceStatistics {
	ist := &NetworkInterfaceStatistics{}
	ist.mtx = &sync.Mutex{}
	return ist
}

func (nist *NetworkInterfaceStatistics) AddTxMessages() {
	nist.mtx.Lock()
	defer nist.mtx.Unlock()
	nist.txMessages++
}

func (nist *NetworkInterfaceStatistics) AddRxMessages() {
	nist.mtx.Lock()
	defer nist.mtx.Unlock()
	nist.rxMessages++
}

func (nist *NetworkInterfaceStatistics) AddTxPackets(data []byte) {
	nist.txPackets++
	nist.txBytes += int64(len(data))
}

func (nist *NetworkInterfaceStatistics) AddRxPackets(data []byte) {
	nist.rxPackets++
	nist.rxBytes += int64(len(data))
}

func (nist *NetworkInterfaceStatistics) AddTxTime(t int64) {
	nist.txTime += t
}

func (nist *NetworkInterfaceStatistics) AddTxTimeSync(t int64) {
	nist.mtx.Lock()
	defer nist.mtx.Unlock()
	nist.txTime += t
}

func (nist *NetworkInterfaceStatistics) String() string {
	buff := &bytes.Buffer{}
	buff.WriteString("Rx Messages:" + strconv.Itoa(int(nist.rxMessages)))
	buff.WriteString(" Tx Messages:" + strconv.Itoa(int(nist.txMessages)))
	buff.WriteString(" Rx Packets:" + strconv.Itoa(int(nist.rxPackets)))
	buff.WriteString(" Tx Packets:" + strconv.Itoa(int(nist.txPackets)))
	buff.WriteString(" Rx Bytes:" + strconv.Itoa(int(nist.rxBytes)))
	buff.WriteString(" Tx Bytes:" + strconv.Itoa(int(nist.txBytes)))
	buff.WriteString(" Avg Tx Speed:" + nist.getTxSpeed())
	return buff.String()
}

func (nist *NetworkInterfaceStatistics) getTxSpeed() string {
	timeFloat := float64(nist.txTime / 1000000000)
	if int64(timeFloat) == 0 {
		//not enought data
		return "N/A"
	}
	speed := float64(nist.txBytes) / timeFloat

	if int64(speed)/1024 == 0 {
		return strconv.Itoa(int(speed)) + " Bytes/Sec"
	}
	speed = speed / 1024
	if int64(speed)/1024 == 0 {
		return strconv.Itoa(int(speed)) + " Kilo Bytes/Sec"
	}
	speed = speed / 1024
	s := strconv.FormatFloat(speed, 'f', 2, 64)
	return s + " Mega Bytes/Sec"
}
