package tests

import (
    "os"
    "testing"

    async_channel "async_channel/async_channel"
    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

func createEmptyChannel(t *testing.T) *channels.Channel {
    channels.DelChan(EmptyTestChannelName)
    ch, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestChannel() },
        channels.SetChan,
        false,
        EmptyTestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    return ch.(*channels.Channel)
}

func TestGetChan(t *testing.T) {
    channels.DelChan(TestChannelName)
    ch := NewEmptyTestChannel()
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    if _, err := channels.GetChan(TestChannelName); err != nil {
        t.Fatalf("get chan: %v", err)
    }
    _ = ch.Stop()
}

func TestSetChan(t *testing.T) {
    channels.DelChan(TestChannelName)
    ch := NewEmptyTestChannel()
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    if _, err := channels.SetChan(NewEmptyTestChannel(), TestChannelName); err == nil {
        t.Fatalf("expected error on duplicate set_chan")
    }
    _ = ch.Stop()
}

func TestSetChanUsingDefaultName(t *testing.T) {
    channels.DelChan(TestChannelName)
    channel := NewEmptyTestChannel()
    channel.Name = "DefaultName"
    returned, err := channels.SetChan(channel, "")
    if err != nil {
        t.Fatalf("set chan: %v", err)
    }
    if returned != channel {
        t.Fatalf("expected same channel")
    }
    if channel.GetName() == "" {
        t.Fatalf("expected channel name")
    }
    if channels.ChannelInstancesInstance().Channels[channel.GetName()] != channel {
        t.Fatalf("channel not stored")
    }
    if _, err := channels.SetChan(NewEmptyTestChannel(), channel.GetName()); err == nil {
        t.Fatalf("expected duplicate error")
    }
    _ = channel.Stop()
}

func TestGetInternalProducer(t *testing.T) {
    channels.DelChan(TestChannelName)
    ch := NewEmptyTestChannel()
    ch.ProducerFactory = nil
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return ch },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    if _, err := ch.GetInternalProducer(); err == nil {
        t.Fatalf("expected error on missing producer class")
    }
    _ = ch.Stop()
}

func TestNewConsumerWithoutProducer(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer, err := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    if len(ch.Consumers) != 1 {
        t.Fatalf("expected 1 consumer")
    }
    _ = consumer
}

func TestNewConsumerWithoutFilters(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer, err := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    if len(ch.GetConsumers()) != 1 || ch.GetConsumers()[0] != consumer {
        t.Fatalf("unexpected consumers list")
    }
}

func TestNewConsumerWithFilters(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer, err := ch.NewConsumer(EmptyTestCallback, map[string]any{"test_key": 1}, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    if len(ch.GetConsumers()) != 1 || ch.GetConsumers()[0] != consumer {
        t.Fatalf("unexpected consumers list")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{})) != 1 {
        t.Fatalf("expected all consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 2})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test2": 2})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1})) != 1 {
        t.Fatalf("expected one consumer")
    }
}

func TestNewConsumerWithExpectedWildcardFilters(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer, err := ch.NewConsumer(EmptyTestCallback, map[string]any{"test_key": 1, "test_key_2": "abc"}, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    if len(ch.GetConsumerFromFilters(map[string]any{})) != 1 {
        t.Fatalf("expected all consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": "abc"})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": "abc", "test_key_3": 45})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": "abc", "test_key_3": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 4, "test_key_2": "bc"})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 3, "test_key_2": async_channel.CHANNEL_WILDCARD})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": async_channel.CHANNEL_WILDCARD, "test_key_2": "abc"})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": async_channel.CHANNEL_WILDCARD, "test_key_2": "a"})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": async_channel.CHANNEL_WILDCARD, "test_key_2": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    _ = consumer
}

func TestNewConsumerWithConsumerWildcardFilters(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer, err := ch.NewConsumer(EmptyTestCallback, map[string]any{"test_key": 1, "test_key_2": "abc", "test_key_3": async_channel.CHANNEL_WILDCARD}, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": "abc"})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": "abc", "test_key_3": 45})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": "abc", "test_key_3": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 4, "test_key_2": "bc"})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1, "test_key_2": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 1})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key_2": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key_3": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key_3": "e"})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": 3, "test_key_2": async_channel.CHANNEL_WILDCARD})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": async_channel.CHANNEL_WILDCARD, "test_key_2": "abc"})) != 1 {
        t.Fatalf("expected one consumer")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": async_channel.CHANNEL_WILDCARD, "test_key_2": "a"})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": async_channel.CHANNEL_WILDCARD, "test_key_2": "a", "test_key_3": async_channel.CHANNEL_WILDCARD})) != 0 {
        t.Fatalf("expected no consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"test_key": async_channel.CHANNEL_WILDCARD, "test_key_2": async_channel.CHANNEL_WILDCARD})) != 1 {
        t.Fatalf("expected one consumer")
    }
    _ = consumer
}

func TestNewConsumerWithMultipleConsumerFiltering(t *testing.T) {
    ch := createEmptyChannel(t)
    consumersDescriptions := []map[string]any{
        {"A": 1, "B": 2, "C": async_channel.CHANNEL_WILDCARD},
        {"A": false, "B": "BBBB", "C": async_channel.CHANNEL_WILDCARD},
        {"A": 3, "B": async_channel.CHANNEL_WILDCARD, "C": async_channel.CHANNEL_WILDCARD},
        {"A": async_channel.CHANNEL_WILDCARD, "B": async_channel.CHANNEL_WILDCARD, "C": async_channel.CHANNEL_WILDCARD},
        {"A": async_channel.CHANNEL_WILDCARD, "B": 2, "C": 1},
        {"A": true, "B": async_channel.CHANNEL_WILDCARD, "C": async_channel.CHANNEL_WILDCARD},
        {"A": nil, "B": nil, "C": async_channel.CHANNEL_WILDCARD},
        {"A": "PPP", "B": 1, "C": async_channel.CHANNEL_WILDCARD, "D": 5},
        {"A": async_channel.CHANNEL_WILDCARD, "B": 2, "C": "ABC"},
        {"A": async_channel.CHANNEL_WILDCARD, "B": true, "C": async_channel.CHANNEL_WILDCARD},
        {"A": async_channel.CHANNEL_WILDCARD, "B": 6, "C": async_channel.CHANNEL_WILDCARD, "D": async_channel.CHANNEL_WILDCARD},
        {"A": async_channel.CHANNEL_WILDCARD, "B": async_channel.CHANNEL_WILDCARD, "C": async_channel.CHANNEL_WILDCARD, "D": async_channel.CHANNEL_WILDCARD},
        {"A": nil, "B": false, "C": "LLLL", "D": async_channel.CHANNEL_WILDCARD},
        {"A": nil, "B": nil, "C": async_channel.CHANNEL_WILDCARD, "D": nil},
        {"A": async_channel.CHANNEL_WILDCARD, "B": 2, "C": async_channel.CHANNEL_WILDCARD, "D": nil},
        {"A": async_channel.CHANNEL_WILDCARD, "B": []any{2, 3, 4, 5, 6}, "C": async_channel.CHANNEL_WILDCARD, "D": nil},
        {"A": async_channel.CHANNEL_WILDCARD, "B": []any{"A", 5, "G"}, "C": async_channel.CHANNEL_WILDCARD, "D": nil},
        {"A": []any{1, 2, 3}, "B": 2, "C": async_channel.CHANNEL_WILDCARD, "D": async_channel.CHANNEL_WILDCARD},
        {"A": []any{"A", "B", "C"}, "B": 2, "C": async_channel.CHANNEL_WILDCARD, "D": async_channel.CHANNEL_WILDCARD},
        {"A": async_channel.CHANNEL_WILDCARD, "B": []any{2}, "C": async_channel.CHANNEL_WILDCARD, "D": async_channel.CHANNEL_WILDCARD},
        {"A": async_channel.CHANNEL_WILDCARD, "B": []any{"B"}, "C": async_channel.CHANNEL_WILDCARD, "D": async_channel.CHANNEL_WILDCARD},
        {"A": 18, "B": []any{"A", "B", "C"}, "C": []any{"---", "9", "#"}, "D": async_channel.CHANNEL_WILDCARD},
        {"A": []any{9, 18}, "B": []any{"B", "C", "D"}, "C": []any{"---", "9", "#", "@", "{"}, "D": []any{"P", "__str__"}},
    }

    var consumers []channels.Consumer
    for _, desc := range consumersDescriptions {
        consumer, err := ch.NewConsumer(EmptyTestCallback, desc, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
        if err != nil {
            t.Fatalf("new consumer: %v", err)
        }
        consumers = append(consumers, consumer)
    }

    if len(ch.GetConsumers()) != len(consumers) {
        t.Fatalf("unexpected consumers count")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{})) != len(consumers) {
        t.Fatalf("expected all consumers")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"A": 1, "B": "6"})) != 3 {
        t.Fatalf("unexpected filter result")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"A": async_channel.CHANNEL_WILDCARD, "B": "G", "C": "1A"})) != 5 {
        t.Fatalf("unexpected filter result")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"A": async_channel.CHANNEL_WILDCARD, "B": async_channel.CHANNEL_WILDCARD, "C": async_channel.CHANNEL_WILDCARD})) != len(consumers) {
        t.Fatalf("unexpected filter result")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"A": 18, "B": "A", "C": "#"})) != 4 {
        t.Fatalf("unexpected filter result")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"A": 18, "B": "C", "C": "#", "D": nil})) != 2 {
        t.Fatalf("unexpected filter result")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"A": 18, "B": "C", "C": "^", "D": nil})) != 1 {
        t.Fatalf("unexpected filter result")
    }
    if len(ch.GetConsumerFromFilters(map[string]any{"A": 18, "B": "C", "C": "#", "D": "__str__"})) != 3 {
        t.Fatalf("unexpected filter result")
    }
}

func TestRemoveConsumer(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer, err := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err != nil {
        t.Fatalf("new consumer: %v", err)
    }
    if len(ch.GetConsumers()) != 1 {
        t.Fatalf("expected 1 consumer")
    }
    _ = ch.RemoveConsumer(consumer)
    if len(ch.GetConsumers()) != 0 {
        t.Fatalf("expected 0 consumer")
    }
}

func TestUnregisterProducer(t *testing.T) {
    ch := createEmptyChannel(t)
    if len(ch.Producers) != 0 {
        t.Fatalf("expected no producers")
    }
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    if len(ch.Producers) != 1 {
        t.Fatalf("expected 1 producer")
    }
}

func TestRegisterProducer(t *testing.T) {
    ch := createEmptyChannel(t)
    if len(ch.Producers) != 0 {
        t.Fatalf("expected no producers")
    }
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    if len(ch.Producers) != 1 {
        t.Fatalf("expected 1 producer")
    }
    ch.UnregisterProducer(producer)
    if len(ch.Producers) != 0 {
        t.Fatalf("expected 0 producers")
    }
}

func TestFlush(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    producer2 := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer2)
    producer3 := NewEmptyTestProducer(ch)
    ch.InternalProducer = producer3

    if producer3.Channel != ch {
        t.Fatalf("expected internal producer channel")
    }
    for _, p := range ch.Producers {
        if p.(*EmptyTestProducer).Channel != ch {
            t.Fatalf("expected producer channel")
        }
    }

    ch.Flush()
    if producer3.Channel != nil {
        t.Fatalf("expected internal producer channel nil")
    }
    for _, p := range ch.Producers {
        if p.(*EmptyTestProducer).Channel != nil {
            t.Fatalf("expected producer channel nil")
        }
    }
}

func TestStart(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer1, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    consumer2, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))

    if err := ch.Start(); err != nil {
        t.Fatalf("start: %v", err)
    }
    _ = consumer1
    _ = consumer2
}

func TestRun(t *testing.T) {
    ch := createEmptyChannel(t)
    consumer1, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    consumer2, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    if err := ch.Run(); err != nil {
        t.Fatalf("run: %v", err)
    }
    _ = consumer1
    _ = consumer2
}

func TestModify(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    producer2 := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer2)
    if err := ch.Modify(map[string]any{}); err != nil {
        t.Fatalf("modify: %v", err)
    }
}

func TestShouldPauseProducersWithNoConsumers(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    ch.IsPaused = false
    if os.Getenv("CYTHON_IGNORE") == "" {
        if !ch.ShouldPauseProducers() {
            t.Fatalf("expected should pause")
        }
    }
}

func TestShouldPauseProducersWhenAlreadyPaused(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    ch.IsPaused = true
    if os.Getenv("CYTHON_IGNORE") == "" {
        if ch.ShouldPauseProducers() {
            t.Fatalf("expected not to pause")
        }
    }
}

func TestShouldPauseProducersWithPriorityConsumers(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    consumer1, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    consumer2, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityMedium))
    consumer3, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    ch.IsPaused = false
    if os.Getenv("CYTHON_IGNORE") == "" {
        if ch.ShouldPauseProducers() {
            t.Fatalf("expected not to pause")
        }
    }
    _ = ch.RemoveConsumer(consumer1)
    _ = ch.RemoveConsumer(consumer2)
    _ = ch.RemoveConsumer(consumer3)
}

func TestShouldPauseProducersWithOptionalConsumers(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    consumer1, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    consumer2, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    consumer3, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    ch.IsPaused = false
    if os.Getenv("CYTHON_IGNORE") == "" {
        if !ch.ShouldPauseProducers() {
            t.Fatalf("expected pause")
        }
    }
    _ = ch.RemoveConsumer(consumer1)
    _ = ch.RemoveConsumer(consumer2)
    _ = ch.RemoveConsumer(consumer3)
}

func TestShouldResumeProducersWithNoConsumers(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    ch.IsPaused = true
    if os.Getenv("CYTHON_IGNORE") == "" {
        if ch.ShouldResumeProducers() {
            t.Fatalf("expected not resume")
        }
    }
}

func TestShouldResumeProducersWhenAlreadyResumed(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    ch.IsPaused = false
    if os.Getenv("CYTHON_IGNORE") == "" {
        if ch.ShouldResumeProducers() {
            t.Fatalf("expected not resume")
        }
    }
}

func TestShouldResumeProducersWithPriorityConsumers(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    consumer1, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityHigh))
    consumer2, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityMedium))
    consumer3, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    ch.IsPaused = true
    if os.Getenv("CYTHON_IGNORE") == "" {
        if !ch.ShouldResumeProducers() {
            t.Fatalf("expected resume")
        }
    }
    _ = ch.RemoveConsumer(consumer1)
    _ = ch.RemoveConsumer(consumer2)
    _ = ch.RemoveConsumer(consumer3)
}

func TestShouldResumeProducersWithOptionalConsumers(t *testing.T) {
    ch := createEmptyChannel(t)
    producer := NewEmptyTestProducer(ch)
    _ = ch.RegisterProducer(producer)
    consumer1, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    consumer2, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    consumer3, _ := ch.NewConsumer(EmptyTestCallback, nil, nil, 0, int(async_channel.ChannelConsumerPriorityOptional))
    ch.IsPaused = true
    if os.Getenv("CYTHON_IGNORE") == "" {
        if ch.ShouldResumeProducers() {
            t.Fatalf("expected not resume")
        }
    }
    _ = ch.RemoveConsumer(consumer1)
    _ = ch.RemoveConsumer(consumer2)
    _ = ch.RemoveConsumer(consumer3)
}
