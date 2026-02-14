package async_channel

// ChannelConsumerPriorityLevels defines consumer priority levels.
type ChannelConsumerPriorityLevels int

const (
    ChannelConsumerPriorityHigh     ChannelConsumerPriorityLevels = 0
    ChannelConsumerPriorityMedium   ChannelConsumerPriorityLevels = 1
    ChannelConsumerPriorityOptional ChannelConsumerPriorityLevels = 2
)
