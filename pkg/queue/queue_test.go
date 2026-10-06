package queue

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
)

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func newTestQueue(t *testing.T, size int) (*Queue, *syncBuffer) {
	t.Helper()
	buf := &syncBuffer{}
	q := New(size, slog.New(slog.NewTextHandler(buf, nil)))
	t.Cleanup(func() { _ = q.Stop(context.Background()) })
	return q, buf
}

func waitFor(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for job")
	}
}

func TestPushAndRun(t *testing.T) {
	q, _ := newTestQueue(t, 10)
	must.NoError(t, q.Start())

	ran := make(chan struct{})
	must.NoError(t, q.Push(func(ctx context.Context) error {
		close(ran)
		return nil
	}))
	waitFor(t, ran)
}

func TestPushNil(t *testing.T) {
	q, _ := newTestQueue(t, 1)
	must.Error(t, q.Push(nil))
	check.Equal(t, q.Len(), 0)
}

func TestFull(t *testing.T) {
	q, _ := newTestQueue(t, 2)
	noop := func(context.Context) error { return nil }

	must.NoError(t, q.Push(noop))
	must.NoError(t, q.Push(noop))
	must.ErrorIs(t, q.Push(noop), ErrFull)
	check.Equal(t, q.Len(), 2)
}

func TestNilLoggerUsesDefault(t *testing.T) {
	q := New(1, nil)
	must.NotNil(t, q.log)
	must.NoError(t, q.Stop(t.Context()))
}

func TestStartIsIdempotentAndSerial(t *testing.T) {
	q, _ := newTestQueue(t, 10)
	must.NoError(t, q.Start())
	must.NoError(t, q.Start())

	started, release, second := make(chan struct{}), make(chan struct{}), make(chan struct{})
	must.NoError(t, q.Push(func(context.Context) error {
		close(started)
		<-release
		return nil
	}))
	waitFor(t, started)
	must.NoError(t, q.Push(func(context.Context) error {
		close(second)
		return nil
	}))

	check.Equal(t, q.Len(), 1) // the only worker is still busy
	close(release)
	waitFor(t, second)
}

func TestStopRejectsNewJobs(t *testing.T) {
	q, _ := newTestQueue(t, 10)
	must.NoError(t, q.Start())
	must.NoError(t, q.Stop(t.Context()))
	must.NoError(t, q.Stop(t.Context()))

	must.ErrorIs(t, q.Push(func(context.Context) error { return nil }), ErrStopped)
	must.ErrorIs(t, q.Start(), ErrStopped)
}

func TestStopBeforeStart(t *testing.T) {
	q, _ := newTestQueue(t, 10)
	must.NoError(t, q.Stop(t.Context()))
	must.ErrorIs(t, q.Start(), ErrStopped)
	must.ErrorIs(t, q.Push(func(context.Context) error { return nil }), ErrStopped)
}

func TestStopWaitsForRunningJobAndDropsPending(t *testing.T) {
	q, buf := newTestQueue(t, 10)
	must.NoError(t, q.Start())

	started := make(chan struct{})
	release := make(chan struct{})
	var pendingRan atomic.Bool

	must.NoError(t, q.Push(func(context.Context) error {
		close(started)
		<-release
		return nil
	}))
	waitFor(t, started)
	for range 2 {
		must.NoError(t, q.Push(func(context.Context) error {
			pendingRan.Store(true)
			return nil
		}))
	}

	stopped := make(chan error, 1)
	go func() { stopped <- q.Stop(context.Background()) }()

	select {
	case <-stopped:
		t.Fatal("Stop returned before running job finished")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	select {
	case err := <-stopped:
		must.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return")
	}

	check.False(t, pendingRan.Load())
	check.Contains(t, buf.String(), "discarding pending jobs")
}

func TestStopTimeoutCancelsJobContext(t *testing.T) {
	q, _ := newTestQueue(t, 10)
	must.NoError(t, q.Start())

	started := make(chan struct{})
	canceled := make(chan struct{})
	must.NoError(t, q.Push(func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		close(canceled)
		return ctx.Err()
	}))
	waitFor(t, started)

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	must.ErrorIs(t, q.Stop(ctx), context.DeadlineExceeded)
	waitFor(t, canceled)
}

func TestJobErrorAndPanicAreLogged(t *testing.T) {
	q, buf := newTestQueue(t, 10)
	must.NoError(t, q.Start())

	done := make(chan struct{})
	must.NoError(t, q.Push(func(context.Context) error { return errors.New("boom") }))
	must.NoError(t, q.Push(func(context.Context) error { panic("kaboom") }))
	must.NoError(t, q.Push(func(context.Context) error {
		close(done)
		return nil
	}))
	waitFor(t, done)

	out := buf.String()
	check.Contains(t, out, "queue job failed")
	check.Contains(t, out, "boom")
	check.Contains(t, out, "queue job panicked")
	check.Contains(t, out, "kaboom")
}
