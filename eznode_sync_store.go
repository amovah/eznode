package eznode

import (
	"time"
)

func (e *EzNode) StartSyncStore() {
	e.store.mutex.Lock()
	defer e.store.mutex.Unlock()

	if !e.store.isRun {
		e.store.ticker = time.NewTicker(e.store.interval)

		go func() {
			for {
				select {
				case <-e.store.done:
					return
				case <-e.store.ticker.C:
					e.store.storage.Save(e.store.errorHandler, e.GetStats())
				}
			}
		}()

		e.store.isRun = true
	}
}

func (e *EzNode) StopSyncStore() {
	e.store.mutex.Lock()
	defer e.store.mutex.Unlock()

	if e.store.isRun {
		e.store.ticker.Stop()
		e.store.done <- true
		e.store.isRun = false
	}
}
