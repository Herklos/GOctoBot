package tests

import (
    "testing"
    "time"

    async_channel "async_channel/async_channel"
    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

func initConsumerTest(t *testing.T) (*channels.Channel, channels.Consumer) {
    channels.DelChan(TestChannelName)
    ch, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestChannel() },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    channel := ch.(*channels.Channel)
    producer := NewEmptyTestProducer(channel)
    if err := producer.Run(); err != nil {
        t.Fatalf("producer run: %v", err)
    }
    consumer, err := channel.NewConsumer(EmptyTestCallback, nil, nil, async_channel.DEFAULT_QUEUE_SIZE, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    return channel, consumer
}

func TestPerformCalled(t *testing.T) {
    channel, consumer := initConsumerTest(t)

    counter := &callCounter{}
    baseConsumer := consumer.(*EmptyTestConsumer)
    baseConsumer.Callback = func(_ map[string]any) error {
        counter.inc()
        return nil
    }

    prod, err := channel.GetInternalProducer()
    if err != nil {
        t.Fatalf("get internal producer: %v", err)
    }
    if err := prod.Send(map[string]any{}); err != nil {
        t.Fatalf("send: %v", err)
    }
    if !waitForCount(counter, 1) {
        t.Fatalf("perform not called")
    }
    _ = channel.Stop()
}

func TestConsumeEndsCalled(t *testing.T) {
    channels.DelChan(TestChannelName)
    ch := NewEmptyTestChannel()
    ch.ConsumerFactory = func(callback channels.Callback, size int, priority int) channels.Consumer {
        return async_channel.NewSupervisedConsumer(callback, size, priority)
    }
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

    consumer, err := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    prod, _ := ch.GetInternalProducer()
    _ = prod.Send(map[string]any{})

    done := make(chan struct{})
    go func() {
        consumer.GetQueue().Join()
        close(done)
    }()

    select {
    case <-done:
    case <-time.After(200 * time.Millisecond):
        t.Fatalf("consume_ends not called")
    }
    _ = ch.Stop()
}

func TestInternalConsumer(t *testing.T) {
    channels.DelChan(TestChannelName)
    ch, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestChannel() },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    channel := ch.(*channels.Channel)
    producer := NewEmptyTestProducer(channel)
    _ = producer.Run()

    internal := async_channel.NewInternalConsumer()
    internal.Callback = func(_ map[string]any) error { return nil }

    _, err = channel.NewConsumer(nil, nil, internal, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new internal consumer: %v", err)
    }

    prod, _ := channel.GetInternalProducer()
    _ = prod.Send(map[string]any{})
    _ = channel.Stop()
}

func TestDefaultInternalConsumerCallback(t *testing.T) {
    internal := async_channel.NewInternalConsumer()
    if err := internal.InternalCallback(map[string]any{}); err == nil {
        t.Fatalf("expected error")
    }
}

func TestSupervisedConsumer(t *testing.T) {
    channels.DelChan(TestChannelName)
    ch := NewEmptyTestChannel()
    ch.ConsumerFactory = func(callback channels.Callback, size int, priority int) channels.Consumer {
        return async_channel.NewSupervisedConsumer(callback, size, priority)
    }
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

    consumer, err := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    prod, _ := ch.GetInternalProducer()
    _ = prod.Send(map[string]any{})
    consumer.GetQueue().Join()
    _ = ch.Stop()
}
