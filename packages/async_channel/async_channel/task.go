package async_channel

import (
    "context"
    "sync"
)

// Task is a lightweight async task wrapper similar to asyncio.Task.
type Task struct {
    ctx    context.Context
    cancel context.CancelFunc
    done   chan struct{}
    once   sync.Once
}

func newTask(fn func(ctx context.Context)) *Task {
    ctx, cancel := context.WithCancel(context.Background())
    t := &Task{
        ctx:    ctx,
        cancel: cancel,
        done:   make(chan struct{}),
    }
    go func() {
        defer t.closeDone()
        fn(ctx)
    }()
    return t
}

func (t *Task) closeDone() {
    t.once.Do(func() {
        close(t.done)
    })
}

// Cancel requests task cancellation.
func (t *Task) Cancel() {
    if t == nil {
        return
    }
    t.cancel()
}

// Done returns a channel closed when the task completes.
func (t *Task) Done() <-chan struct{} {
    if t == nil {
        ch := make(chan struct{})
        close(ch)
        return ch
    }
    return t.done
}

// Context returns the task context.
func (t *Task) Context() context.Context {
    if t == nil {
        return context.Background()
    }
    return t.ctx
}
