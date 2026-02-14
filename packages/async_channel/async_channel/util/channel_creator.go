package util

import "async_channel/async_channel/channels"

// ChannelFactory creates a channel instance.
type ChannelFactory func() channels.ChannelInterface

var subclassRegistry = map[string][]ChannelFactory{}

// RegisterSubclass registers a channel factory for a base class name.
func RegisterSubclass(baseName string, factory ChannelFactory) {
    subclassRegistry[baseName] = append(subclassRegistry[baseName], factory)
}

// CreateAllSubclassesChannel creates instances for registered subclasses.
func CreateAllSubclassesChannel(
    baseName string,
    setChanMethod func(ch channels.ChannelInterface, name string) (channels.ChannelInterface, error),
    isSynchronized bool,
) error {
    for _, factory := range subclassRegistry[baseName] {
        if _, err := CreateChannelInstance(factory, setChanMethod, isSynchronized, ""); err != nil {
            return err
        }
    }
    return nil
}

// CreateChannelInstance creates, initializes, and starts a channel instance.
func CreateChannelInstance(
    factory ChannelFactory,
    setChanMethod func(ch channels.ChannelInterface, name string) (channels.ChannelInterface, error),
    isSynchronized bool,
    channelName string,
) (channels.ChannelInterface, error) {
    created := factory()
    if _, err := setChanMethod(created, channelName); err != nil {
        return nil, err
    }
    created.SetSynchronized(isSynchronized)
    if err := created.Start(); err != nil {
        return nil, err
    }
    return created, nil
}
