package tests

import (
	. "github.com/saichler/messaging/golang/net/protocol"
	. "github.com/saichler/utils/golang"
	"os"
	"strconv"
	"testing"
	"time"
)

func setup() {
	//SetLevel(DebugLevel)
}
func tearDown() {}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestNetworkNode(t *testing.T) {
	NetConfig.Set(512, false)
	replyCount = 0
	h1 := NewStringMessageHandler()
	h2 := NewStringMessageHandler()

	Info("Node1:", h1.node.ServiceID().String(), " Node2:", h2.node.ServiceID().String())

	time.Sleep(time.Second * 2)

	h1.SendString("Hello World", h2.node.ServiceID())
	h2.SendString("Hello World", h1.node.ServiceID())

	time.Sleep(time.Second * 2)

	if replyCount != 2 {
		t.Fail()
		Error("Expected 2 and got " + strconv.Itoa(replyCount))
	}

	h1.node.Shutdown()
	h2.node.Shutdown()
	time.Sleep(time.Second * 2)
}

func TestSwitch(t *testing.T) {
	replyCount = 0
	NetConfig.Set(512, false)
	s := NewStringMessageHandler()
	h1 := NewStringMessageHandler()
	h2 := NewStringMessageHandler()

	Info("Node1:", h1.node.ServiceID().String(), " Node2:", h2.node.ServiceID().String())

	time.Sleep(time.Second * 2)

	h1.SendString("Hello World", h2.node.ServiceID())
	h2.SendString("Hello World", h1.node.ServiceID())

	time.Sleep(time.Second * 2)

	if replyCount != 2 {
		t.Fail()
		Error("Expected 2 and got " + strconv.Itoa(replyCount))
	}

	h1.node.Shutdown()
	h2.node.Shutdown()
	s.node.Shutdown()
	time.Sleep(time.Second * 2)
}

func TestMultiPart(t *testing.T) {
	replyCount = 0
	NetConfig.Set(4, false)
	h1 := NewStringMessageHandler()
	h2 := NewStringMessageHandler()

	Info("Node1:", h1.node.ServiceID().String(), " Node2:", h2.node.ServiceID().String())

	time.Sleep(time.Second * 2)

	h1.SendString("Hello World", h2.node.ServiceID())
	h2.SendString("Hello World", h1.node.ServiceID())

	time.Sleep(time.Second * 2)

	if replyCount != 2 {
		t.Fail()
		Error("Expected 2 and got " + strconv.Itoa(replyCount))
	}

	h1.node.Shutdown()
	h2.node.Shutdown()
	time.Sleep(time.Second * 2)
}

func TestMessageScale(t *testing.T) {
	replyCount = 0
	NetConfig.Set(512, false)
	numOfMessages := 10000
	numOfNodes := 3

	handlers := make([]*StringMessageHandler, numOfNodes)
	for i := 0; i < len(handlers); i++ {
		h := NewStringMessageHandler()
		h.print = false
		handlers[i] = h
		Info("Habitat HabitatID:" + handlers[i].node.ServiceID().String())
	}

	time.Sleep(time.Second * 2)
	for i := 1; i < len(handlers)-1; i++ {
		go sendScale(handlers[i], handlers[i+1], numOfMessages)
	}

	time.Sleep(time.Second * 2)

	if replyCount != numOfMessages {
		t.Fail()
		Error("Expected " + strconv.Itoa(numOfMessages) + " and got " + strconv.Itoa(replyCount))
	} else {
		Info("Passed sending & receiving " + strconv.Itoa(replyCount) + " messages")
	}

	for _, h := range handlers {
		h.node.Shutdown()
	}
	time.Sleep(time.Second * 2)
}

func sendScale(h1, h2 *StringMessageHandler, size int) {
	for i := 0; i < size; i++ {
		h1.SendString("Hello World:"+strconv.Itoa(i), h2.node.ServiceID())
	}
}

func TestNetworkNodeAndMessageScale(t *testing.T) {
	replyCount = 0
	NetConfig.Set(512, false)
	numOfMessages := 10000
	numOfNodes := 50

	handlers := make([]*StringMessageHandler, numOfNodes)
	for i := 0; i < len(handlers); i++ {
		h := NewStringMessageHandler()
		h.print = false
		handlers[i] = h
		Info("Habitat HabitatID:" + handlers[i].node.ServiceID().String())
	}

	time.Sleep(time.Second * 2)
	for i := 1; i < len(handlers)-1; i++ {
		go sendScale(handlers[i], handlers[i+1], numOfMessages)
	}

	time.Sleep(time.Second * 10)

	if replyCount != numOfMessages*(len(handlers)-2) {
		t.Fail()
		Error("Expected " + strconv.Itoa(numOfMessages*(len(handlers)-2)) + " and got " + strconv.Itoa(replyCount))
	} else {
		Info("Passed sending & receiving " + strconv.Itoa(replyCount) + " messages")
	}

	for _, h := range handlers {
		h.node.Shutdown()
	}
	time.Sleep(time.Second * 2)
}

func TestPublish(t *testing.T) {
	NetConfig.Set(512, false)
	publishID := NewServiceID(NetConfig.PublishID(), "publish", 0)
	replyCount = 0
	numOfNodes := 50

	handlers := make([]*StringMessageHandler, numOfNodes)
	for i := 0; i < len(handlers); i++ {
		h := NewStringMessageHandler()
		h.print = false
		handlers[i] = h
		Info("Habitat HabitatID:" + handlers[i].node.ServiceID().String())
	}

	time.Sleep(time.Second * 2)

	handlers[2].SendString("Hello World Multicast", publishID)

	time.Sleep(time.Second * 3)

	if replyCount != len(handlers) {
		t.Fail()
		Error("Expected " + strconv.Itoa(len(handlers)) + " and got " + strconv.Itoa(replyCount))
	} else {
		Info("Passed sending & receiving " + strconv.Itoa(replyCount) + " messages")
	}

	for _, h := range handlers {
		h.node.Shutdown()
	}
	time.Sleep(time.Second * 2)
}

func TestNetworkNodeEncrypted(t *testing.T) {
	NetConfig.Set(512, true)
	replyCount = 0
	h1 := NewStringMessageHandler()
	h2 := NewStringMessageHandler()

	Info("Node1:", h1.node.ServiceID().String(), " Node2:", h2.node.ServiceID().String())

	time.Sleep(time.Second * 2)

	h1.SendString("Hello World", h2.node.ServiceID())
	h2.SendString("Hello World", h1.node.ServiceID())

	time.Sleep(time.Second * 2)

	if replyCount != 2 {
		t.Fail()
		Error("Expected 2 and got " + strconv.Itoa(replyCount))
	}
	h1.node.Shutdown()
	h2.node.Shutdown()
	time.Sleep(time.Second * 2)
}

func TestNetworkNodeAndMessageScaleSecure(t *testing.T) {
	replyCount = 0
	NetConfig.Set(512, true)
	numOfMessages := 10000
	numOfNodes := 50

	handlers := make([]*StringMessageHandler, numOfNodes)
	for i := 0; i < len(handlers); i++ {
		h := NewStringMessageHandler()
		h.print = false
		handlers[i] = h
		Info("Habitat HabitatID:" + handlers[i].node.ServiceID().String())
	}

	time.Sleep(time.Second * 6)

	for i := 1; i < len(handlers)-1; i++ {
		go sendScale(handlers[i], handlers[i+1], numOfMessages)
	}

	time.Sleep(time.Second * 10)

	if replyCount != numOfMessages*(numOfNodes-2) {
		t.Fail()
		Error("Expected " + strconv.Itoa(numOfMessages*(numOfNodes-2)) + " and got " + strconv.Itoa(replyCount))
	} else {
		Info("Passed sending & receiving " + strconv.Itoa(replyCount) + " messages")
	}

	for _, h := range handlers {
		h.node.Shutdown()
	}
	time.Sleep(time.Second * 2)
}

func TestDestinationUnreachable(t *testing.T) {
	unreachCount = 0
	replyCount = 0
	NetConfig.Set(512, false)
	h1 := NewStringMessageHandler()
	h2 := NewStringMessageHandler()

	Info("Node1:", h1.node.ServiceID().String(), " Node2:", h2.node.ServiceID().String())

	time.Sleep(time.Second * 2)

	unreachInternal := NewServiceID(NewNetworkID(GetLocalIpAddress(), 52005), "", 0)

	h1.SendString("Hello World", unreachInternal)
	h2.SendString("Hello World", unreachInternal)

	time.Sleep(time.Second * 2)

	if unreachCount != 2 {
		t.Fail()
		Error("Expected 2 and got " + strconv.Itoa(unreachCount))
	}
	h1.node.Shutdown()
	h2.node.Shutdown()
	time.Sleep(time.Second * 2)
}

func TestMultiPartUnreachable(t *testing.T) {
	unreachCount = 0
	replyCount = 0
	NetConfig.Set(4, false)
	h1 := NewStringMessageHandler()
	h2 := NewStringMessageHandler()

	time.Sleep(time.Second * 2)
	unreachInternal := NewServiceID(NewNetworkID(GetLocalIpAddress(), 52005), "", 0)
	h2.SendString("Hello World", unreachInternal)

	time.Sleep(time.Second * 2)

	if unreachCount != 1 {
		t.Fail()
		Error("Expected 1 and got " + strconv.Itoa(unreachCount))
	}
	h1.node.Shutdown()
	h2.node.Shutdown()
	time.Sleep(time.Second * 2)
}

func TestShutdown(t *testing.T) {
	unreachCount = 0
	replyCount = 0
	NetConfig.Set(4, false)
	h1 := NewStringMessageHandler()
	h2 := NewStringMessageHandler()

	time.Sleep(time.Second * 3)

	h1.node.Shutdown()
	h2.node.Shutdown()

	time.Sleep(time.Second * 5)
}
