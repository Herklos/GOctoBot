package tests

import (
    "sync"
    "time"

    async_channel "async_channel/async_channel"
    "async_channel/async_channel/channels"
)

const (
    TestChannelName         = "Test"
    EmptyTestChannelName    = "EmptyTest"
    EmptyTestWithIDName     = "EmptyTestWithId"
    ConsumerKey             = "test"
)

type EmptyTestConsumer struct {
    *async_channel.Consumer
}

func NewEmptyTestConsumer(callback channels.Callback, size int, priority int) channels.Consumer {
    return &EmptyTestConsumer{Consumer: async_channel.NewConsumer(callback, size, priority)}
}

type EmptyTestSupervisedConsumer struct {
    *async_channel.SupervisedConsumer
}

func NewEmptyTestSupervisedConsumer(callback channels.Callback, size int, priority int) channels.Consumer {
    return &EmptyTestSupervisedConsumer{SupervisedConsumer: async_channel.NewSupervisedConsumer(callback, size, priority)}
}

type EmptyTestProducer struct {
    *async_channel.Producer
}

func NewEmptyTestProducer(ch *channels.Channel) *EmptyTestProducer {
    return &EmptyTestProducer{Producer: async_channel.NewProducer(ch)}
}

func (p *EmptyTestProducer) Start() error {
    time.Sleep(100 * time.Millisecond)
    return nil
}

func (p *EmptyTestProducer) Pause() error { return nil }
func (p *EmptyTestProducer) Resume() error { return nil }

func NewEmptyTestChannel() *channels.Channel {
    ch := channels.NewChannel()
    ch.Name = EmptyTestChannelName
    ch.ConsumerFactory = func(callback channels.Callback, size int, priority int) channels.Consumer {
        return NewEmptyTestConsumer(callback, size, priority)
    }
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return NewEmptyTestProducer(c)
    }
    return ch
}

func NewEmptyTestWithIDChannel(id string) *channels.Channel {
    ch := NewEmptyTestChannel()
    ch.Name = EmptyTestWithIDName
    ch.ChanID = id
    return ch
}

func EmptyTestCallback(_ map[string]any) error { return nil }

func waitNextCycle() {
    time.Sleep(10 * time.Millisecond)
}

type callCounter struct {
    mu    sync.Mutex
    count int
}

func (c *callCounter) inc() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}

func (c *callCounter) value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

func waitForCount(counter *callCounter, expected int) bool {
    deadline := time.Now().Add(200 * time.Millisecond)
    for time.Now().Before(deadline) {
        if counter.value() == expected {
            return true
        }
        time.Sleep(5 * time.Millisecond)
    }
    return counter.value() == expected
}

func resetChannelInstances() {
    channels.ChannelInstancesInstance().Channels = map[string]any{}
}
