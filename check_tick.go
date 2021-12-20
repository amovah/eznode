package eznode

import "time"

type CheckTick struct {
	TickRate         time.Duration
	MaxCheckDuration time.Duration
}
