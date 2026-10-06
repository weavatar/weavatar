// Package queue is an in-process job queue drained by a single worker.
package queue

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"sync"
)

var (
	ErrFull    = errors.New("queue: full")
	ErrStopped = errors.New("queue: stopped")
)

// Job receives a ctx canceled when Stop times out; long jobs should watch it.
type Job func(ctx context.Context) error

// Queue runs jobs one at a time, in push order.
type Queue struct {
	jobs chan Job
	log  *slog.Logger

	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.RWMutex
	started  bool
	stopped  bool
	quit     chan struct{} // closed by Stop; the worker takes no more jobs
	quitOnce sync.Once
	done     chan struct{} // closed when the worker exits
}

// New buffers size jobs; with size <= 0 a push only succeeds while the worker
// is idle and waiting.
func New(size int, log *slog.Logger) *Queue {
	if log == nil {
		log = slog.Default()
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		jobs:   make(chan Job, max(size, 0)),
		log:    log,
		ctx:    ctx,
		cancel: cancel,
		quit:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

// Push never blocks: it returns ErrFull or ErrStopped instead.
func (q *Queue) Push(job Job) error {
	if job == nil {
		return errors.New("queue: nil job")
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.stopped {
		return ErrStopped
	}

	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrFull
	}
}

// Start is idempotent; after Stop it returns ErrStopped.
func (q *Queue) Start() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return ErrStopped
	}
	if q.started {
		return nil
	}

	q.started = true
	go q.run()

	return nil
}

// Stop refuses new jobs, drops pending ones and waits for the running job,
// canceling its ctx if ctx expires first.
func (q *Queue) Stop(ctx context.Context) error {
	q.mu.Lock()
	q.stopped = true
	started := q.started
	q.mu.Unlock()

	q.quitOnce.Do(func() { close(q.quit) })
	defer q.cancel()

	if !started {
		return nil
	}

	select {
	case <-q.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Len counts jobs not yet started.
func (q *Queue) Len() int {
	return len(q.jobs)
}

func (q *Queue) run() {
	defer close(q.done)

	for {
		select {
		case <-q.quit:
			q.discardPending()
			return
		case job := <-q.jobs:
			// select picks randomly when quit and jobs are both ready
			select {
			case <-q.quit:
				q.discardPending()
				return
			default:
			}
			q.process(job)
		}
	}
}

func (q *Queue) process(job Job) {
	defer func() {
		if r := recover(); r != nil {
			q.log.Error("queue job panicked", slog.Any("panic", r), slog.String("stack", string(debug.Stack())))
		}
	}()

	if err := job(q.ctx); err != nil {
		q.log.Error("queue job failed", slog.Any("err", err))
	}
}

func (q *Queue) discardPending() {
	if n := len(q.jobs); n > 0 {
		q.log.Warn("queue stopped, discarding pending jobs", slog.Int("count", n))
	}
}
