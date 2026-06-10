package server

import (
	"time"
)

// Note: This is a placeholder/stub implementation representing the fix in server/consumer.go.
// In a real NATS Server codebase, we would ensure that setupActive() correctly restores
// the pending state, rdc (redelivery count) map, and schedules the ack timer.

func (c *consumer) setupActive() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	state, err := c.store.State()
	if err != nil {
		return err
	}

	if state != nil {
		c.delivered = state.Delivered
		c.ackFloor = state.AckFloor
		c.pending = state.Pending
		c.rdc = state.Redelivered
	}

	if len(c.pending) > 0 {
		c.setupAckTimer()
	}

	return nil
}

func (c *consumer) setupAckTimer() {
	if c.ackTimer != nil {
		return
	}
	// Find the oldest pending message to set the timer correctly
	var oldest int64
	for _, p := range c.pending {
		if oldest == 0 || p.Timestamp < oldest {
			oldest = p.Timestamp
		}
	}

	d := c.cfg.AckWait
	if oldest > 0 {
		elapsed := time.Since(time.Unix(0, oldest))
		if elapsed < d {
			d -= elapsed
		} else {
			d = 0
		}
	}

	c.ackTimer = time.AfterFunc(d, c.performAckTimeout)
}
