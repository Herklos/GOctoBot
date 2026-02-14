package tests

import (
    "testing"
    "time"

    async_channel "async_channel/async_channel"
    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

const synchronizedChannelName = "SynchronizedTest"

type SynchronizedProducerTest struct {
    *async_channel.Producer
    channel *channels.Channel
}

func (p *SynchronizedProducerTest) Send(data any) error {
    if err := p.Producer.Send(data); err != nil {
        return err
    }
    _ = p.channel.Stop()
    return nil
}

func (p *SynchronizedProducerTest) Pause() error { return nil }
func (p *SynchronizedProducerTest) Resume() error { return nil }

func newSynchronizedProducer(ch *channels.Channel) *SynchronizedProducerTest {
    return &SynchronizedProducerTest{Producer: async_channel.NewProducer(ch), channel: ch}
}

func newSynchronizedChannel() *channels.Channel {
    ch := NewEmptyTestChannel()
    ch.Name = synchronizedChannelName
    ch.ProducerFactory = func(c *channels.Channel) channels.Producer {
        return newSynchronizedProducer(c)
    }
    ch.ConsumerFactory = func(callback channels.Callback, size int, priority int) channels.Consumer {
        return NewEmptyTestConsumer(callback, size, priority)
    }
    return ch
}

func setupSynchronizedChannel(t *testing.T) *channels.Channel {
    channels.DelChan(synchronizedChannelName)
    ch := newSynchronizedChannel()
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        true,
        synchronizedChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    return ch
}

func TestProducerSynchronizedPerformConsumersQueueWithOneConsumer(t *testing.T) {
    ch := setupSynchronizedChannel(t)

    callbackCount := &callCounter{}
    consumer, _ := ch.NewConsumer(func(_ map[string]any) error {
        callbackCount.inc()
        return nil
    }, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))

    producer := newSynchronizedProducer(ch)
    _ = producer.Run()

    _ = producer.Send(map[string]any{})
    if callbackCount.value() != 0 {
        t.Fatalf("callback should not be called yet")
    }
    _ = producer.SynchronizedPerformConsumersQueue(1, true, 50*time.Millisecond)
    if callbackCount.value() != 1 {
        t.Fatalf("callback should be called once")
    }
    _ = consumer
}

func TestProducerSynchronizedPerformSupervisedConsumerWithProcessingEmptyQueue(t *testing.T) {
    ch := setupSynchronizedChannel(t)

    continueCh := make(chan struct{})
    calls := &callCounter{}
    doneCalls := &callCounter{}

    callback := func(_ map[string]any) error {
        calls.inc()
        <-continueCh
        doneCalls.inc()
        return nil
    }

    ch.ConsumerFactory = func(cb channels.Callback, size int, priority int) channels.Consumer {
        return async_channel.NewSupervisedConsumer(cb, size, priority)
    }
    consumer, _ := ch.NewConsumer(callback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))

    producer := newSynchronizedProducer(ch)
    _ = producer.Run()

    _ = producer.Send(map[string]any{})
    _ = consumer.Run(true)

    waitNextCycle()
    if calls.value() != 1 {
        t.Fatalf("expected callback called")
    }
    if doneCalls.value() != 0 {
        t.Fatalf("expected callback not finished")
    }
    if consumer.GetQueue().QSize() != 0 {
        t.Fatalf("expected empty queue")
    }

    close(continueCh)

    _ = producer.SynchronizedPerformConsumersQueue(1, false, 50*time.Millisecond)
    if doneCalls.value() != 0 {
        t.Fatalf("expected not done without join")
    }
    _ = producer.SynchronizedPerformConsumersQueue(1, true, 50*time.Millisecond)
    if doneCalls.value() != 1 {
        t.Fatalf("expected done after join")
    }
    _ = consumer.Stop()
}

func TestJoin(t *testing.T) {
    baseConsumer := async_channel.NewConsumer(nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    _ = baseConsumer.Join(50 * time.Millisecond)

    supervised := async_channel.NewSupervisedConsumer(nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if supervised.IsIdleState() == false {
        t.Fatalf("expected idle")
    }
    _ = supervised.Join(50 * time.Millisecond)
    supervised.SetIdleState(false)
    go func() {
        time.Sleep(10 * time.Millisecond)
        supervised.SetIdleState(true)
    }()
    if err := supervised.Join(50 * time.Millisecond); err != nil {
        t.Fatalf("expected join to wait")
    }
}

func TestJoinQueue(t *testing.T) {
    baseConsumer := async_channel.NewConsumer(nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    _ = baseConsumer.JoinQueue()

    supervised := async_channel.NewSupervisedConsumer(nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    _ = supervised.JoinQueue()
}

func TestSynchronizedNoTasks(t *testing.T) {
    ch := setupSynchronizedChannel(t)

    consumer, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    producer := newSynchronizedProducer(ch)
    _ = producer.Run()

    if consumer.(*EmptyTestConsumer).ConsumeTask != nil {
        t.Fatalf("expected no consume task")
    }
    if producer.ProduceTask != nil {
        t.Fatalf("expected no produce task")
    }
}

func TestIsConsumersQueueEmptyWithOneConsumer(t *testing.T) {
    ch := setupSynchronizedChannel(t)

    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    producer := newSynchronizedProducer(ch)
    _ = producer.Run()

    _ = producer.Send(map[string]any{})
    if producer.IsConsumersQueueEmpty(1) {
        t.Fatalf("expected not empty")
    }
    if producer.IsConsumersQueueEmpty(2) {
        t.Fatalf("expected not empty")
    }
    _ = producer.SynchronizedPerformConsumersQueue(1, true, 50*time.Millisecond)
    if !producer.IsConsumersQueueEmpty(1) {
        t.Fatalf("expected empty")
    }
    if !producer.IsConsumersQueueEmpty(2) {
        t.Fatalf("expected empty")
    }
}

func TestIsConsumersQueueEmptyWithMultipleConsumers(t *testing.T) {
    ch := setupSynchronizedChannel(t)

    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, 1)
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, 1)
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, 2)
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, 2)
    _, _ = ch.NewConsumer(EmptyTestCallback, nil, nil, 0, 3)

    producer := newSynchronizedProducer(ch)
    _ = producer.Run()

    _ = producer.Send(map[string]any{})
    if producer.IsConsumersQueueEmpty(1) || producer.IsConsumersQueueEmpty(2) || producer.IsConsumersQueueEmpty(3) {
        t.Fatalf("expected queues not empty")
    }
    _ = producer.SynchronizedPerformConsumersQueue(1, true, 50*time.Millisecond)
    if !producer.IsConsumersQueueEmpty(1) {
        t.Fatalf("expected empty for priority 1")
    }
    if producer.IsConsumersQueueEmpty(2) {
        t.Fatalf("expected not empty for priority 2")
    }
    if producer.IsConsumersQueueEmpty(3) {
        t.Fatalf("expected not empty for priority 3")
    }
    _ = producer.SynchronizedPerformConsumersQueue(2, true, 50*time.Millisecond)
    if !producer.IsConsumersQueueEmpty(2) {
        t.Fatalf("expected empty for priority 2")
    }
    if producer.IsConsumersQueueEmpty(3) {
        t.Fatalf("expected not empty for priority 3")
    }
    _ = producer.SynchronizedPerformConsumersQueue(3, true, 50*time.Millisecond)
    if !producer.IsConsumersQueueEmpty(3) {
        t.Fatalf("expected empty for priority 3")
    }
}

func TestProducerSynchronizedPerformConsumersQueueWithMultipleConsumer(t *testing.T) {
    ch := setupSynchronizedChannel(t)

    counters := []*callCounter{&callCounter{}, &callCounter{}, &callCounter{}, &callCounter{}, &callCounter{}}
    _, _ = ch.NewConsumer(func(_ map[string]any) error { counters[0].inc(); return nil }, nil, nil, 0, 1)
    _, _ = ch.NewConsumer(func(_ map[string]any) error { counters[1].inc(); return nil }, nil, nil, 0, 1)
    _, _ = ch.NewConsumer(func(_ map[string]any) error { counters[2].inc(); return nil }, nil, nil, 0, 2)
    _, _ = ch.NewConsumer(func(_ map[string]any) error { counters[3].inc(); return nil }, nil, nil, 0, 2)
    _, _ = ch.NewConsumer(func(_ map[string]any) error { counters[4].inc(); return nil }, nil, nil, 0, 3)

    producer := newSynchronizedProducer(ch)
    _ = producer.Run()

    _ = producer.Send(map[string]any{})
    for _, c := range counters {
        if c.value() != 0 {
            t.Fatalf("expected no callbacks before sync perform")
        }
    }
    _ = producer.SynchronizedPerformConsumersQueue(1, true, 50*time.Millisecond)
    if counters[0].value() != 1 || counters[1].value() != 1 {
        t.Fatalf("expected priority 1 callbacks")
    }
    if counters[2].value() != 0 || counters[3].value() != 0 || counters[4].value() != 0 {
        t.Fatalf("unexpected callbacks")
    }
    _ = producer.SynchronizedPerformConsumersQueue(2, true, 50*time.Millisecond)
    if counters[2].value() != 1 || counters[3].value() != 1 {
        t.Fatalf("expected priority 2 callbacks")
    }
    if counters[4].value() != 0 {
        t.Fatalf("unexpected priority 3 callback")
    }
    if producer.IsConsumersQueueEmpty(3) {
        t.Fatalf("expected not empty before priority 3")
    }
    _ = producer.SynchronizedPerformConsumersQueue(3, true, 50*time.Millisecond)
    if counters[4].value() != 1 {
        t.Fatalf("expected priority 3 callbacks")
    }
    if !producer.IsConsumersQueueEmpty(1) || !producer.IsConsumersQueueEmpty(2) || !producer.IsConsumersQueueEmpty(3) {
        t.Fatalf("expected all queues empty")
    }

    _ = producer.Send(map[string]any{})
    _ = producer.SynchronizedPerformConsumersQueue(3, true, 50*time.Millisecond)
    for _, c := range counters {
        if c.value() != 2 {
            t.Fatalf("expected callbacks called twice")
        }
    }
}
