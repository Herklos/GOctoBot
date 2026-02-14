package tests

import (
    "fmt"
    "sync/atomic"
    "testing"

    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

var idCounter uint64

func newID() string {
    return fmt.Sprintf("id-%d", atomic.AddUint64(&idCounter, 1))
}

func TestGetChanAtID(t *testing.T) {
    resetChannelInstances()
    channelID := newID()
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestWithIDChannel(channelID) },
        channels.SetChanAtID,
        false,
        EmptyTestWithIDName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    if _, err := channels.GetChanAtID(EmptyTestWithIDName, channelID); err != nil {
        t.Fatalf("get chan at id: %v", err)
    }
}

func TestSetChanAtIDAlreadyExist(t *testing.T) {
    resetChannelInstances()
    channelID := newID()
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestWithIDChannel(channelID) },
        channels.SetChanAtID,
        false,
        EmptyTestWithIDName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    if _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestWithIDChannel(channelID) },
        channels.SetChanAtID,
        false,
        EmptyTestWithIDName,
    ); err == nil {
        t.Fatalf("expected error")
    }
}

func TestDelChannelContainerNotExistDoesNotRaise(t *testing.T) {
    resetChannelInstances()
    channelID := newID()
    channels.DelChannelContainer(channelID + "test")
}

func TestDelChannelContainer(t *testing.T) {
    resetChannelInstances()
    channelID := newID()
    _, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestWithIDChannel(channelID) },
        channels.SetChanAtID,
        false,
        EmptyTestWithIDName,
    )
    if err != nil {
        t.Fatalf("create channel instance: %v", err)
    }
    channels.DelChannelContainer(channelID)
    if _, err := channels.GetChanAtID(EmptyTestWithIDName, channelID); err == nil {
        t.Fatalf("expected error")
    }
    channels.DelChanAtID(EmptyTestWithIDName, channelID)
}

func TestGetChannelsNotExist(t *testing.T) {
    resetChannelInstances()
    channelID := newID()
    if _, err := channels.GetChannels(channelID + "test"); err == nil {
        t.Fatalf("expected error")
    }
}

func TestGetChannels(t *testing.T) {
    resetChannelInstances()
    chanID := newID()
    channel4ID := newID()
    channel6ID := newID()

    ch1, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { return NewEmptyTestWithIDChannel(chanID) },
        channels.SetChanAtID,
        false,
        EmptyTestWithIDName,
    )
    if err != nil {
        t.Fatalf("create channel 1: %v", err)
    }
    ch2, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { c := NewEmptyTestWithIDChannel(chanID); c.Name = "EmptyTestWithId2"; return c },
        channels.SetChanAtID,
        false,
        "EmptyTestWithId2",
    )
    if err != nil {
        t.Fatalf("create channel 2: %v", err)
    }
    ch3, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { c := NewEmptyTestWithIDChannel(chanID); c.Name = "EmptyTestWithId3"; return c },
        channels.SetChanAtID,
        false,
        "EmptyTestWithId3",
    )
    if err != nil {
        t.Fatalf("create channel 3: %v", err)
    }
    ch4, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { c := NewEmptyTestWithIDChannel(channel4ID); c.Name = "EmptyTestWithId4"; return c },
        channels.SetChanAtID,
        false,
        "EmptyTestWithId4",
    )
    if err != nil {
        t.Fatalf("create channel 4: %v", err)
    }
    ch5, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { c := NewEmptyTestWithIDChannel(channel4ID); c.Name = "EmptyTestWithId5"; return c },
        channels.SetChanAtID,
        false,
        "EmptyTestWithId5",
    )
    if err != nil {
        t.Fatalf("create channel 5: %v", err)
    }
    ch6, err := util.CreateChannelInstance(
        func() channels.ChannelInterface { c := NewEmptyTestWithIDChannel(channel6ID); c.Name = "EmptyTestWithId6"; return c },
        channels.SetChanAtID,
        false,
        "EmptyTestWithId6",
    )
    if err != nil {
        t.Fatalf("create channel 6: %v", err)
    }

    if channels, _ := channels.GetChannels(chanID); len(channels) != 3 {
        t.Fatalf("expected 3 channels")
    }
    if channels, _ := channels.GetChannels(channel4ID); len(channels) != 2 {
        t.Fatalf("expected 2 channels")
    }
    if channels, _ := channels.GetChannels(channel6ID); len(channels) != 1 {
        t.Fatalf("expected 1 channel")
    }

    _ = ch1
    _ = ch2
    _ = ch3
    _ = ch4
    _ = ch5
    _ = ch6

    channels.DelChannelContainer(chanID)
    channels.DelChannelContainer(channel4ID)
    channels.DelChannelContainer(channel6ID)
}
