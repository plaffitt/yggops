package time

import (
	"context"
	"time"
)

type TriggerableTicker struct {
	C <-chan time.Time
	c chan time.Time
}

func NewTriggerableTicker(d time.Duration, ctx context.Context) *TriggerableTicker {
	t := &TriggerableTicker{}
	ticker := time.NewTicker(d)
	t.c = make(chan time.Time, 1)
	t.C = t.c

	go func() {
		for {
			select {
			case <-ticker.C:
				t.TriggerUpdate()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	return t
}

func (t *TriggerableTicker) TriggerUpdate() {
	select {
	case t.c <- time.Now():
	default:
	}
}
