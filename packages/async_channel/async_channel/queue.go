package async_channel

import (
    "errors"
    "sync"
)

var (
    ErrQueueClosed = errors.New("queue closed")
    ErrTaskDone    = errors.New("task_done called too many times")
)

// Queue mirrors asyncio.Queue semantics with optional max size (0 = unlimited).
type Queue struct {
    maxSize   int
    items     []any
    unfinished int
    closed    bool

    mu       sync.Mutex
    notEmpty *sync.Cond
    notFull  *sync.Cond
    allDone  *sync.Cond
}

func NewQueue(maxSize int) *Queue {
    q := &Queue{maxSize: maxSize}
    q.notEmpty = sync.NewCond(&q.mu)
    q.notFull = sync.NewCond(&q.mu)
    q.allDone = sync.NewCond(&q.mu)
    return q
}

func (q *Queue) Close() {
    q.mu.Lock()
    q.closed = true
    q.notEmpty.Broadcast()
    q.notFull.Broadcast()
    q.allDone.Broadcast()
    q.mu.Unlock()
}

// WakeAll wakes up any goroutines waiting on the queue.
func (q *Queue) WakeAll() {
    q.mu.Lock()
    q.notEmpty.Broadcast()
    q.notFull.Broadcast()
    q.allDone.Broadcast()
    q.mu.Unlock()
}

func (q *Queue) Put(item any) error {
    q.mu.Lock()
    defer q.mu.Unlock()
    for !q.closed && q.maxSize > 0 && len(q.items) >= q.maxSize {
        q.notFull.Wait()
    }
    if q.closed {
        return ErrQueueClosed
    }
    q.items = append(q.items, item)
    q.unfinished++
    q.notEmpty.Signal()
    return nil
}

func (q *Queue) Get(cancel <-chan struct{}) (any, error) {
    q.mu.Lock()
    defer q.mu.Unlock()
    for !q.closed && len(q.items) == 0 {
        if cancel != nil {
            select {
            case <-cancel:
                return nil, ErrQueueClosed
            default:
            }
        }
        q.notEmpty.Wait()
    }
    if q.closed && len(q.items) == 0 {
        return nil, ErrQueueClosed
    }
    item := q.items[0]
    q.items = q.items[1:]
    q.notFull.Signal()
    return item, nil
}

func (q *Queue) TaskDone() error {
    q.mu.Lock()
    defer q.mu.Unlock()
    if q.unfinished == 0 {
        return ErrTaskDone
    }
    q.unfinished--
    if q.unfinished == 0 {
        q.allDone.Broadcast()
    }
    return nil
}

func (q *Queue) Join() {
    q.mu.Lock()
    defer q.mu.Unlock()
    for q.unfinished > 0 && !q.closed {
        q.allDone.Wait()
    }
}

func (q *Queue) Empty() bool {
    q.mu.Lock()
    defer q.mu.Unlock()
    return len(q.items) == 0
}

func (q *Queue) QSize() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    return len(q.items)
}
