package channels

import (
    "fmt"
    "log"
    "sync"
)

// ChannelInstances singleton.
type ChannelInstances struct {
    Channels map[string]any
}

var (
    instancesMu sync.Mutex
    instances   = map[*ChannelInstances]struct{}{}
    singleton   *ChannelInstances
)

// ChannelInstancesInstance returns the singleton instance.
func ChannelInstancesInstance() *ChannelInstances {
    instancesMu.Lock()
    defer instancesMu.Unlock()
    if singleton == nil {
        singleton = &ChannelInstances{Channels: map[string]any{}}
        instances[singleton] = struct{}{}
    }
    return singleton
}

func SetChanAtID(chanInst ChannelInterface, name string) (ChannelInterface, error) {
    chanName := chanInst.GetName()
    if name != "" {
        chanName = name
    }
    chanID := ""
    if c, ok := chanInst.(*Channel); ok {
        chanID = c.ChanID
    }
    if chanID == "" {
        return nil, fmt.Errorf("chan_id is empty")
    }

    containerAny, ok := ChannelInstancesInstance().Channels[chanID]
    var container map[string]ChannelInterface
    if !ok {
        container = map[string]ChannelInterface{}
        ChannelInstancesInstance().Channels[chanID] = container
    } else {
        existing, ok := containerAny.(map[string]ChannelInterface)
        if !ok {
            return nil, fmt.Errorf("chan_id container invalid")
        }
        container = existing
    }

    if _, exists := container[chanName]; !exists {
        container[chanName] = chanInst
        return chanInst, nil
    }
    return nil, fmt.Errorf("Channel %s already exists.", chanName)
}

func GetChannels(chanID string) (map[string]ChannelInterface, error) {
    containerAny, ok := ChannelInstancesInstance().Channels[chanID]
    if !ok {
        return nil, fmt.Errorf("Channels not found with chan_id: %s", chanID)
    }
    container, ok := containerAny.(map[string]ChannelInterface)
    if !ok {
        return nil, fmt.Errorf("Channels not found with chan_id: %s", chanID)
    }
    return container, nil
}

func DelChannelContainer(chanID string) {
    delete(ChannelInstancesInstance().Channels, chanID)
}

func GetChanAtID(chanName string, chanID string) (ChannelInterface, error) {
    container, err := GetChannels(chanID)
    if err != nil {
        return nil, fmt.Errorf("Channel %s not found with chan_id: %s", chanName, chanID)
    }
    chanInst, ok := container[chanName]
    if !ok {
        return nil, fmt.Errorf("Channel %s not found with chan_id: %s", chanName, chanID)
    }
    return chanInst, nil
}

func DelChanAtID(chanName string, chanID string) {
    containerAny, ok := ChannelInstancesInstance().Channels[chanID]
    if !ok {
        log.Printf("Can't del chan %s with chan_id: %s", chanName, chanID)
        return
    }
    container, ok := containerAny.(map[string]ChannelInterface)
    if !ok {
        log.Printf("Can't del chan %s with chan_id: %s", chanName, chanID)
        return
    }
    delete(container, chanName)
}
