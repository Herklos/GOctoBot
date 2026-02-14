package tests

import (
    "testing"

    async_channel "async_channel/async_channel"
    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

type TestProducer struct {
    *async_channel.Producer
    channel *channels.Channel
}

func (p *TestProducer) Send(data any) error {
    if err := p.Producer.Send(data); err != nil {
        return err
    }
    _ = p.channel.Stop()
    return nil
}

func (p *TestProducer) Pause() error { return nil }
func (p *TestProducer) Resume() error { return nil }

func newTestProducer(ch *channels.Channel) *TestProducer {
    return &TestProducer{Producer: async_channel.NewProducer(ch), channel: ch}
}

func TestSendInternalProducerWithoutConsumer(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return newTestProducer(c)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    prod, err := ch.GetInternalProducer()
    if err != nil {
        t.Fatalf("get internal producer: %v", err)
    }
    _ = prod.Send(map[string]any{})
}

func TestSendProducerWithoutConsumer(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return newTestProducer(c)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }

    producer := newTestProducer(ch)
    _ = producer.Run()
    _ = producer.Send(map[string]any{})
}

func TestSendProducerWithConsumer(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ConsumerFactory = func(callback channels.Callback, size int, priority int) channels.Consumer {
        return async_channel.NewConsumer(callback, size, priority)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }

    callback := func(data map[string]any) error {
        if data["data"] != "test" {
            t.Fatalf("unexpected data")
        }
        _ = ch.Stop()
        return nil
    }
    _, err = ch.NewConsumer(callback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }

    producer := NewEmptyTestProducer(ch)
    _ = producer.Run()
    _ = producer.Send(map[string]any{"data": "test"})
}

func TestPauseProducerWithoutConsumers(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return newTestProducer(c)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    _ = newTestProducer(ch).Run()
}

func TestPauseProducerWithRemovedConsumer(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return newTestProducer(c)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    consumer, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    _ = newTestProducer(ch).Run()
    _ = ch.RemoveConsumer(consumer)
}

func TestResumeProducer(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return newTestProducer(c)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    _ = newTestProducer(ch).Run()
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
}

func TestWaitForProcessing(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ConsumerFactory = func(callback channels.Callback, size int, priority int) channels.Consumer {
        return async_channel.NewSupervisedConsumer(callback, size, priority)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    producer := NewEmptyTestProducer(ch)
    _ = producer.Run()
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    _ = producer.Send(map[string]any{"data": "test"})
    _ = producer.WaitForProcessing()
    _ = ch.Stop()
}

func TestProducerIsRunning(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    producer := NewEmptyTestProducer(ch)
    if producer.IsRunning {
        t.Fatalf("expected not running")
    }
    _ = producer.Run()
    if !producer.IsRunning {
        t.Fatalf("expected running")
    }
    _ = ch.Stop()
    if producer.IsRunning {
        t.Fatalf("expected not running")
    }
}

func TestProducerPauseResume(t *testing.T) {
    ch := NewEmptyTestChannel()
    ch.Name = TestChannelName
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return async_channel.NewProducer(c)
    }
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    producer := async_channel.NewProducer(ch)
    if !producer.Channel.IsPaused {
        t.Fatalf("expected paused")
    }
    _ = producer.Pause()
    if !producer.Channel.IsPaused {
        t.Fatalf("expected paused")
    }
    _ = producer.Resume()
    if producer.Channel.IsPaused {
        t.Fatalf("expected resumed")
    }
    _ = producer.Pause()
    if !producer.Channel.IsPaused {
        t.Fatalf("expected paused")
    }
    _ = producer.Resume()
    if producer.Channel.IsPaused {
        t.Fatalf("expected resumed")
    }
    _ = ch.Stop()
}
