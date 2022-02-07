package eznode

import "time"

// CheckTick determines interval of checking nodes availability
type CheckTick struct {
	// TickRate is the interval of checking nodes availability
	TickRate time.Duration
	// MaxCheckDuration is the maximum duration of checking nodes availability
	MaxCheckDuration time.Duration
}
