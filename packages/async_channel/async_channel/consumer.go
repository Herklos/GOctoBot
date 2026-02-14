package async_channel

import (
    "context"
    "fmt"
    "sync"
    "time"

    "async_channel/async_channel/channels"
    "async_channel/async_channel/util"
)

// Callback mirrors an async callback with kwargs-like map.
type Callback = channels.Callback

// Consumer keeps reading from the channel and processes data.
type Consumer struct {
    Logger        util.Logger
    Queue         *Queue
    Callback      Callback
    ConsumeTask   *Task
    ShouldStop    bool
    PriorityLevelValue int
    performFunc      func(any) error
    consumeEndsFunc  func() error
    stopCh        chan struct{}
    stopChClosed  bool
}

func NewConsumer(callback Callback, size int, priorityLevel int) *Consumer {
    c := &Consumer{
        Logger:        util.GetLogger("Consumer"),
        Queue:         NewQueue(size),
        Callback:      callback,
        ConsumeTask:   nil,
        ShouldStop:    false,
        PriorityLevelValue: priorityLevel,
        stopCh:       make(chan struct{}),
        stopChClosed: false,
    }
    return c
}

// Consume reads from queue and processes data until stopped.
func (c *Consumer) Consume() {
    for !c.ShouldStop {
        item, err := c.Queue.Get(c.stopCh)
        if err == ErrQueueClosed {
            return
        }
        if err != nil {
            c.Logger.Errorf("consume error: %v", err)
            continue
        }
        if err := c.Perform(item); err != nil {
            c.Logger.Errorf("Exception when calling callback on %v: %v", c, err)
        }
        if err := c.ConsumeEnds(); err != nil {
            if err != ErrTaskDone {
                c.Logger.Errorf("consume_ends error: %v", err)
            }
        }
    }
}

// Perform handles queue data.
func (c *Consumer) Perform(data any) error {
    if c.performFunc != nil {
        return c.performFunc(data)
    }
    return c.performDefault(data)
}

func (c *Consumer) performDefault(data any) error {
    if c.Callback == nil {
        return nil
    }
    if data == nil {
        return c.Callback(map[string]any{})
    }
    if kwargs, ok := data.(map[string]any); ok {
        return c.Callback(kwargs)
    }
    return c.Callback(map[string]any{"data": data})
}

// ConsumeEnds handles consumption end.
func (c *Consumer) ConsumeEnds() error {
    if c.consumeEndsFunc != nil {
        return c.consumeEndsFunc()
    }
    return nil
}

// Start initializes consumer.
func (c *Consumer) Start() error {
    c.ShouldStop = false
    if c.stopChClosed {
        c.stopCh = make(chan struct{})
        c.stopChClosed = false
    }
    return nil
}

// Stop stops consumer tasks.
func (c *Consumer) Stop() error {
    c.ShouldStop = true
    if c.ConsumeTask != nil {
        c.ConsumeTask.Cancel()
    }
    if !c.stopChClosed {
        close(c.stopCh)
        c.stopChClosed = true
    }
    c.Queue.WakeAll()
    return nil
}

// CreateTask starts consume in a task.
func (c *Consumer) CreateTask() {
    c.ConsumeTask = newTask(func(ctx context.Context) {
        _ = ctx
        c.Consume()
    })
}

// Run initializes and starts consuming.
func (c *Consumer) Run(withTask bool) error {
    if err := c.Start(); err != nil {
        return err
    }
    if withTask {
        c.CreateTask()
    }
    return nil
}

// Join is a no-op for base consumer.
func (c *Consumer) Join(timeout time.Duration) error {
    return nil
}

// JoinQueue is a no-op for base consumer.
func (c *Consumer) JoinQueue() error {
    return nil
}

// PriorityLevel returns the consumer priority level.
func (c *Consumer) PriorityLevel() int {
    return c.PriorityLevelValue
}

// GetQueue returns the consumer queue.
func (c *Consumer) GetQueue() channels.Queue {
    return c.Queue
}

func (c *Consumer) String() string {
    if c.Callback == nil {
        return fmt.Sprintf("%T with callback: <nil>", c)
    }
    return fmt.Sprintf("%T with callback", c)
}

// InternalConsumer uses internal callback.
type InternalConsumer struct {
    *Consumer
}

func NewInternalConsumer() *InternalConsumer {
    ic := &InternalConsumer{}
    ic.Consumer = NewConsumer(nil, DEFAULT_QUEUE_SIZE, int(ChannelConsumerPriorityHigh))
    ic.Callback = ic.InternalCallback
    return ic
}

func (c *InternalConsumer) InternalCallback(_ map[string]any) error {
    return fmt.Errorf("internal_callback is not implemented")
}

// SupervisedConsumer notifies queue when work is done.
type SupervisedConsumer struct {
    *Consumer
    idleMu sync.Mutex
    idle   *sync.Cond
    isIdle bool
}

func NewSupervisedConsumer(callback Callback, size int, priorityLevel int) *SupervisedConsumer {
    sc := &SupervisedConsumer{
        Consumer: NewConsumer(callback, size, priorityLevel),
        isIdle:   true,
    }
    sc.idle = sync.NewCond(&sc.idleMu)
    sc.Consumer.performFunc = func(data any) error {
        sc.idleMu.Lock()
        sc.isIdle = false
        sc.idleMu.Unlock()
        defer func() {
            sc.idleMu.Lock()
            sc.isIdle = true
            sc.idleMu.Unlock()
            sc.idle.Broadcast()
        }()
        return sc.Consumer.performDefault(data)
    }
    sc.Consumer.consumeEndsFunc = func() error {
        return sc.Queue.TaskDone()
    }
    return sc
}

func (c *SupervisedConsumer) Join(timeout time.Duration) error {
    c.idleMu.Lock()
    if c.isIdle {
        c.idleMu.Unlock()
        return nil
    }
    c.idleMu.Unlock()

    done := make(chan struct{})
    go func() {
        c.idleMu.Lock()
        for !c.isIdle {
            c.idle.Wait()
        }
        c.idleMu.Unlock()
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("join timeout")
    }
}

func (c *SupervisedConsumer) JoinQueue() error {
    c.Queue.Join()
    return nil
}

// Perform and ConsumeEnds are handled via Consumer function hooks.

// SetIdleState allows tests to control idle state.
func (c *SupervisedConsumer) SetIdleState(value bool) {
    c.idleMu.Lock()
    c.isIdle = value
    c.idleMu.Unlock()
    c.idle.Broadcast()
}

// IsIdleState returns current idle state.
func (c *SupervisedConsumer) IsIdleState() bool {
    c.idleMu.Lock()
    defer c.idleMu.Unlock()
    return c.isIdle
}
