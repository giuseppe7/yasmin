package workers

import (
	"testing"
	"time"
)

func TestDoWork(t *testing.T) {
	worker, err := NewSlackWorker()
	if err != nil {
		t.Errorf("unexpected error with invoking NewSlackWorker")
		return
	}
	go worker.DoWork()
	// TODO: I hate this.
	time.Sleep(15 * time.Second)
	// If we get here, no errors so all must be good.
}
