package eznode

import "time"

type CheckTick struct {
	tickRate         time.Duration
	maxCheckDuration time.Duration
}

func NewCheckTick(tickRate time.Duration, maxCheckDuration time.Duration) CheckTick {
	return CheckTick{
		tickRate:         tickRate,
		maxCheckDuration: maxCheckDuration,
	}
}
