package tests

import (
    "testing"

    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

func TestCreateChannelInstance(t *testing.T) {
    resetChannelInstances()
    channels.DelChan(TestChannelName)
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestChannel() },
        channels.SetChan,
        false,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    ch, _ := channels.GetChan(TestChannelName)
    _ = ch.(*channels.Channel).Stop()
}

func TestCreateSynchronizedChannelInstance(t *testing.T) {
    resetChannelInstances()
    channels.DelChan(TestChannelName)
    ch, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestChannel() },
        channels.SetChan,
        true,
        TestChannelName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    channel := ch.(*channels.Channel)
    if !channel.IsSynchronized {
        t.Fatalf("expected synchronized")
    }
    _ = channel.Stop()
}

func TestCreateAllSubclassesChannel(t *testing.T) {
    resetChannelInstances()
    baseName := "TestChannelClass"

    cleanChannels := func() {
        for key := range channels.ChannelInstancesInstance().Channels {
            channels.DelChan(key)
        }
    }

    util.RegisterSubclass(baseName, func() channels.ChannelInterface {
        ch := NewEmptyTestChannel()
        ch.Name = "Test1"
        return ch
    })
    util.RegisterSubclass(baseName, func() channels.ChannelInterface {
        ch := NewEmptyTestChannel()
        ch.Name = "Test2"
        return ch
    })

    cleanChannels()
    if err := util.CreateAllSubclassesChannel(baseName, channels.SetChan, false); err != nil {
        t.Fatalf("create subclasses: %v", err)
    }

    names := []string{}
    for name := range channels.ChannelInstancesInstance().Channels {
        names = append(names, name)
    }
    if len(names) != 2 {
        t.Fatalf("expected 2 channels")
    }

    cleanChannels()
    if err := util.CreateAllSubclassesChannel(baseName, channels.SetChan, true); err != nil {
        t.Fatalf("create subclasses: %v", err)
    }
    syncChannels := channels.ChannelInstancesInstance().Channels
    if len(syncChannels) != 2 {
        t.Fatalf("expected 2 channels")
    }
    for _, value := range syncChannels {
        if ch, ok := value.(channels.ChannelInterface); ok {
            if c, ok := ch.(*channels.Channel); ok {
                if !c.IsSynchronized {
                    t.Fatalf("expected synchronized")
                }
            }
        }
    }
    cleanChannels()
}
