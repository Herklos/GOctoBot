package channels

import (
    "fmt"
    "reflect"
    "strings"
    "time"
)

// Callback mirrors async callback with kwargs-like map.
type Callback func(map[string]any) error

// Queue defines the queue interface used by channels.
type Queue interface {
    Put(item any) error
    Get(cancel <-chan struct{}) (any, error)
    TaskDone() error
    Join()
    Empty() bool
    QSize() int
}

// Consumer defines the consumer interface used by channels.
type Consumer interface {
    Start() error
    Stop() error
    Run(withTask bool) error
    Join(timeout time.Duration) error
    JoinQueue() error
    Perform(data any) error
    PriorityLevel() int
    GetQueue() Queue
}

// Producer defines the producer interface used by channels.
type Producer interface {
    Send(data any) error
    Pause() error
    Resume() error
    Stop() error
    Modify(kwargs map[string]any) error
}

// ChannelInterface defines the public channel interface.
type ChannelInterface interface {
    GetName() string
    Start() error
    Stop() error
    Run() error
    Modify(kwargs map[string]any) error
    SetSynchronized(value bool)
}

// Channel is the core async channel type.
type Channel struct {
    Name string
    ChanID string

    Producers []Producer
    Consumers []map[string]any

    InternalProducer Producer

    IsPaused       bool
    IsSynchronized bool

    ProducerFactory func(ch *Channel) Producer
    ConsumerFactory func(callback Callback, size int, priorityLevel int) Consumer
}

const (
    InstanceKey          = "consumer_instance"
    DefaultPriorityLevel = 0
    ChannelWildcard      = "*"
    PriorityOptional     = 2
)

func NewChannel() *Channel {
    return &Channel{
        Producers:     []Producer{},
        Consumers:     []map[string]any{},
        InternalProducer: nil,
        IsPaused:      true,
        IsSynchronized: false,
    }
}

func (c *Channel) GetName() string {
    if c.Name != "" {
        return c.Name
    }
    typ := reflect.TypeOf(c)
    if typ.Kind() == reflect.Ptr {
        typ = typ.Elem()
    }
    name := typ.Name()
    return strings.TrimSuffix(name, "Channel")
}

// NewConsumer creates and registers a consumer.
func (c *Channel) NewConsumer(
    callback Callback,
    consumerFilters map[string]any,
    internalConsumer Consumer,
    size int,
    priorityLevel int,
) (Consumer, error) {
    var consumer Consumer
    if internalConsumer != nil {
        consumer = internalConsumer
    } else {
        if c.ConsumerFactory == nil {
            return nil, fmt.Errorf("CONSUMER_CLASS not defined")
        }
        consumer = c.ConsumerFactory(callback, size, priorityLevel)
    }
    if err := c.addNewConsumerAndRun(consumer, consumerFilters); err != nil {
        return nil, err
    }
    if err := c.checkProducersState(); err != nil {
        return nil, err
    }
    return consumer, nil
}

func (c *Channel) addNewConsumerAndRun(consumer Consumer, consumerFilters map[string]any) error {
    if consumerFilters == nil {
        consumerFilters = map[string]any{}
    }
    c.AddNewConsumer(consumer, consumerFilters)
    return consumer.Run(!c.IsSynchronized)
}

func (c *Channel) AddNewConsumer(consumer Consumer, consumerFilters map[string]any) {
    consumerFilters[InstanceKey] = consumer
    c.Consumers = append(c.Consumers, consumerFilters)
}

func (c *Channel) GetConsumerFromFilters(consumerFilters map[string]any) []Consumer {
    return c.filterConsumers(consumerFilters)
}

func (c *Channel) GetConsumers() []Consumer {
    out := make([]Consumer, 0, len(c.Consumers))
    for _, consumer := range c.Consumers {
        if instance, ok := consumer[InstanceKey].(Consumer); ok {
            out = append(out, instance)
        }
    }
    return out
}

func (c *Channel) GetPrioritizedConsumers(priorityLevel int) []Consumer {
    out := []Consumer{}
    for _, consumer := range c.Consumers {
        instance, ok := consumer[InstanceKey].(Consumer)
        if !ok {
            continue
        }
        if instance.PriorityLevel() <= priorityLevel {
            out = append(out, instance)
        }
    }
    return out
}

func (c *Channel) filterConsumers(consumerFilters map[string]any) []Consumer {
    out := []Consumer{}
    for _, consumer := range c.Consumers {
        if checkFilters(consumer, consumerFilters) {
            if instance, ok := consumer[InstanceKey].(Consumer); ok {
                out = append(out, instance)
            }
        }
    }
    return out
}

func (c *Channel) RemoveConsumer(consumer Consumer) error {
    for idx, consumerCandidate := range c.Consumers {
        if instance, ok := consumerCandidate[InstanceKey].(Consumer); ok && instance == consumer {
            c.Consumers = append(c.Consumers[:idx], c.Consumers[idx+1:]...)
            if err := c.checkProducersState(); err != nil {
                return err
            }
            return consumer.Stop()
        }
    }
    return nil
}

func (c *Channel) checkProducersState() error {
    if c.ShouldPauseProducers() {
        c.IsPaused = true
        for _, producer := range c.GetProducers() {
            if err := producer.Pause(); err != nil {
                return err
            }
        }
        return nil
    }
    if c.ShouldResumeProducers() {
        c.IsPaused = false
        for _, producer := range c.GetProducers() {
            if err := producer.Resume(); err != nil {
                return err
            }
        }
    }
    return nil
}

func (c *Channel) ShouldPauseProducers() bool {
    if c.IsPaused {
        return false
    }
    if len(c.GetConsumers()) == 0 {
        return true
    }
    for _, consumer := range c.GetConsumers() {
        if consumer.PriorityLevel() < PriorityOptional {
            return false
        }
    }
    return true
}

func (c *Channel) ShouldResumeProducers() bool {
    if !c.IsPaused {
        return false
    }
    if len(c.GetConsumers()) == 0 {
        return false
    }
    for _, consumer := range c.GetConsumers() {
        if consumer.PriorityLevel() < PriorityOptional {
            return true
        }
    }
    return false
}

func (c *Channel) RegisterProducer(producer Producer) error {
    exists := false
    for _, p := range c.Producers {
        if p == producer {
            exists = true
            break
        }
    }
    if !exists {
        c.Producers = append(c.Producers, producer)
    }
    if c.IsPaused {
        return producer.Pause()
    }
    return nil
}

func (c *Channel) UnregisterProducer(producer Producer) {
    for i, p := range c.Producers {
        if p == producer {
            c.Producers = append(c.Producers[:i], c.Producers[i+1:]...)
            return
        }
    }
}

func (c *Channel) GetProducers() []Producer {
    return c.Producers
}

func (c *Channel) Start() error {
    for _, consumer := range c.GetConsumers() {
        if err := consumer.Start(); err != nil {
            return err
        }
    }
    return nil
}

func (c *Channel) Stop() error {
    for _, consumer := range c.GetConsumers() {
        _ = consumer.Stop()
    }
    for _, producer := range c.GetProducers() {
        _ = producer.Stop()
    }
    if c.InternalProducer != nil {
        _ = c.InternalProducer.Stop()
    }
    return nil
}

func (c *Channel) Flush() {
    if c.InternalProducer != nil {
        if p, ok := c.InternalProducer.(interface{ SetChannel(*Channel) }); ok {
            p.SetChannel(nil)
        }
    }
    for _, producer := range c.GetProducers() {
        if p, ok := producer.(interface{ SetChannel(*Channel) }); ok {
            p.SetChannel(nil)
        }
    }
}

func (c *Channel) Run() error {
    for _, consumer := range c.GetConsumers() {
        if err := consumer.Run(!c.IsSynchronized); err != nil {
            return err
        }
    }
    return nil
}

func (c *Channel) Modify(kwargs map[string]any) error {
    for _, producer := range c.GetProducers() {
        if err := producer.Modify(kwargs); err != nil {
            return err
        }
    }
    return nil
}

func (c *Channel) GetInternalProducer() (Producer, error) {
    if c.InternalProducer == nil {
        if c.ProducerFactory == nil {
            return nil, fmt.Errorf("PRODUCER_CLASS not defined")
        }
        c.InternalProducer = c.ProducerFactory(c)
    }
    return c.InternalProducer, nil
}

func (c *Channel) SetSynchronized(value bool) {
    c.IsSynchronized = value
}

func SetChan(chanInst ChannelInterface, name string) (ChannelInterface, error) {
    chanName := name
    if chanName == "" {
        chanName = chanInst.GetName()
    }
    if _, exists := ChannelInstancesInstance().Channels[chanName]; !exists {
        ChannelInstancesInstance().Channels[chanName] = chanInst
        return chanInst, nil
    }
    return nil, fmt.Errorf("Channel %s already exists.", chanName)
}

func DelChan(name string) {
    delete(ChannelInstancesInstance().Channels, name)
}

func GetChan(chanName string) (ChannelInterface, error) {
    value, ok := ChannelInstancesInstance().Channels[chanName]
    if !ok {
        return nil, fmt.Errorf("Channel %s not found", chanName)
    }
    chanInst, ok := value.(ChannelInterface)
    if !ok {
        return nil, fmt.Errorf("Channel %s has invalid type", chanName)
    }
    return chanInst, nil
}

func checkFilters(consumerFilters map[string]any, expectedFilters map[string]any) bool {
    if len(expectedFilters) == 0 {
        return true
    }
    for key, value := range expectedFilters {
        if value == ChannelWildcard {
            continue
        }
        actual, ok := consumerFilters[key]
        if !ok {
            return false
        }
        if isSlice(actual) {
            if listContains(actual, value) || listContains(actual, ChannelWildcard) {
                continue
            }
            return false
        }
        if !valuesEqual(actual, value) && !valuesEqual(actual, ChannelWildcard) {
            return false
        }
    }
    return true
}

func isSlice(value any) bool {
    v := reflect.ValueOf(value)
    if v.Kind() == reflect.Slice {
        return true
    }
    return false
}

func listContains(list any, target any) bool {
    v := reflect.ValueOf(list)
    if v.Kind() != reflect.Slice {
        return false
    }
    for i := 0; i < v.Len(); i++ {
        if valuesEqual(v.Index(i).Interface(), target) {
            return true
        }
    }
    return false
}

func valuesEqual(actual any, expected any) bool {
    if actual == nil && expected == nil {
        return true
    }
    if actual == nil || expected == nil {
        return false
    }

    // Handle bool/int/float equivalence (Python True == 1, False == 0)
    if ab, ok := actual.(bool); ok {
        if num, okNum := numericValue(expected); okNum {
            return (ab && num == 1) || (!ab && num == 0)
        }
    }
    if eb, ok := expected.(bool); ok {
        if num, okNum := numericValue(actual); okNum {
            return (eb && num == 1) || (!eb && num == 0)
        }
    }

    return reflect.DeepEqual(actual, expected)
}

func numericValue(value any) (float64, bool) {
    switch v := value.(type) {
    case int:
        return float64(v), true
    case int8:
        return float64(v), true
    case int16:
        return float64(v), true
    case int32:
        return float64(v), true
    case int64:
        return float64(v), true
    case uint:
        return float64(v), true
    case uint8:
        return float64(v), true
    case uint16:
        return float64(v), true
    case uint32:
        return float64(v), true
    case uint64:
        return float64(v), true
    case float32:
        return float64(v), true
    case float64:
        return v, true
    default:
        return 0, false
    }
}
