package regression

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jb843051627/avalanche-net/internal/ingest"
)

type boomErr struct{}

func (boomErr) Error() string { return "boom" }

func TestBug08_SubmitAndRetryHonorCanceledContext(t *testing.T) {
	q := ingest.New(2)
	block := make(chan struct{})
	job := ingest.Job{Run: func(context.Context) error { <-block; return nil }}
	for i := 0; i < 3; i++ {
		if err := q.Submit(context.Background(), job); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	errc := make(chan error, 1)
	go func() {
		errc <- q.Submit(ctx, ingest.Job{Run: func(context.Context) error { return nil }})
	}()
	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("submit canceled err=%v, want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("submit ignored canceled context (blocked forever)")
	}
	close(block)
	if err := ingest.Retry(ctx, 20, func() error { return boomErr{} }); !errors.Is(err, context.Canceled) {
		t.Fatalf("retry canceled err=%v, want context.Canceled", err)
	}
	q.Close()
}
