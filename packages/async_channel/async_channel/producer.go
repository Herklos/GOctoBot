package async_channel

import (
    "context"
    "time"

    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

// Producer produces output to consumers.
type Producer struct {
    Logger      util.Logger
    Channel     *channels.Channel
    ProduceTask *Task
    ShouldStop  bool
    IsRunning   bool
}

func NewProducer(channel *channels.Channel) *Producer {
    return &Producer{
        Logger:    util.GetLogger("Producer"),
        Channel:   channel,
        ShouldStop: false,
        IsRunning: false,
    }
}

// Send enqueues data to each consumer queue.
func (p *Producer) Send(data any) error {
    for _, consumer := range p.Channel.GetConsumers() {
        if err := consumer.GetQueue().Put(data); err != nil {
            return err
        }
    }
    return nil
}

// Push notifies that new data should be sent.
func (p *Producer) Push(_ map[string]any) error {
    return nil
}

// Start implements producer's non-triggered tasks.
func (p *Producer) Start() error {
    return nil
}

// Pause is called when channel runs out of consumers.
func (p *Producer) Pause() error {
    p.Logger.Debugf("Pausing...")
    p.IsRunning = false
    if !p.Channel.IsPaused {
        p.Channel.IsPaused = true
    }
    return nil
}

// Resume is called when channel has consumers.
func (p *Producer) Resume() error {
    p.Logger.Debugf("Resuming...")
    if p.Channel.IsPaused {
        p.Channel.IsPaused = false
    }
    return nil
}

// Perform implements producer's non-triggered tasks.
func (p *Producer) Perform(_ map[string]any) error {
    return nil
}

// Modify is called when producer can be modified during perform.
func (p *Producer) Modify(_ map[string]any) error {
    return nil
}

// WaitForProcessing waits until all consumers queues are processed.
func (p *Producer) WaitForProcessing() error {
    for _, consumer := range p.Channel.GetConsumers() {
        _ = consumer.JoinQueue()
    }
    return nil
}

// SynchronizedPerformConsumersQueue empties queues synchronously for each consumer.
func (p *Producer) SynchronizedPerformConsumersQueue(priorityLevel int, joinConsumers bool, timeout time.Duration) error {
    for _, consumer := range p.Channel.GetPrioritizedConsumers(priorityLevel) {
        for !consumer.GetQueue().Empty() {
            item, err := consumer.GetQueue().Get(nil)
            if err != nil {
                return err
            }
            if err := consumer.Perform(item); err != nil {
                return err
            }
        }
        if joinConsumers {
            if err := consumer.Join(timeout); err != nil {
                return err
            }
        }
    }
    return nil
}

// Stop stops producer tasks.
func (p *Producer) Stop() error {
    p.ShouldStop = true
    p.IsRunning = false
    if p.ProduceTask != nil {
        p.ProduceTask.Cancel()
    }
    return nil
}

// CreateTask starts a new task running Start().
func (p *Producer) CreateTask() {
    p.IsRunning = true
    p.ProduceTask = newTask(func(ctx context.Context) {
        _ = p.Start()
    })
}

// Run starts producer main task and registers it to channel.
func (p *Producer) Run() error {
    if err := p.Channel.RegisterProducer(p); err != nil {
        return err
    }
    if !p.Channel.IsSynchronized {
        p.CreateTask()
    }
    return nil
}

// IsConsumersQueueEmpty checks if consumers queues are empty for given priority level.
func (p *Producer) IsConsumersQueueEmpty(priorityLevel int) bool {
    for _, consumer := range p.Channel.GetConsumers() {
        if consumer.PriorityLevel() <= priorityLevel && !consumer.GetQueue().Empty() {
            return false
        }
    }
    return true
}

// SetChannel allows channel to be detached on flush.
func (p *Producer) SetChannel(ch *channels.Channel) {
    p.Channel = ch
}
