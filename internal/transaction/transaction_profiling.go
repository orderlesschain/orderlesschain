package transaction

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sync"
	"time"
)

type LatencyBreakDown struct {
	SequenceStart time.Time
	SequenceEnd   time.Time
	OrderStart    time.Time
	OrderEnd      time.Time
	EndorseStart  time.Time
	EndorseEnd    time.Time
	ExecBidlStart time.Time
	ExecBidlEnd   time.Time
	CommitStart   time.Time
	CommitEnd     time.Time
}

type TransactionsProfiling struct {
	TransactionsProfilingLock *sync.Mutex
	TransactionsProfiling     map[string]*LatencyBreakDown
}

func InitTransactionsProfiling() *TransactionsProfiling {
	return &TransactionsProfiling{
		TransactionsProfilingLock: &sync.Mutex{},
		TransactionsProfiling:     map[string]*LatencyBreakDown{},
	}
}

func (t *TransactionsProfiling) PrepareBreakdownToBeSent(transactionId string, latencyBreakDown *LatencyBreakDown) *protos.LatencyBreakDown {
	latencies := &protos.LatencyBreakDown{
		TransactionId: transactionId,
	}
	if !(latencyBreakDown.SequenceStart.IsZero() || latencyBreakDown.SequenceEnd.IsZero()) {
		latencies.SequenceDuration = int64(latencyBreakDown.SequenceEnd.Sub(latencyBreakDown.SequenceStart))
	}
	if !(latencyBreakDown.OrderStart.IsZero() || latencyBreakDown.OrderEnd.IsZero()) {
		latencies.OrderDuration = int64(latencyBreakDown.OrderEnd.Sub(latencyBreakDown.OrderStart))
	}
	if !(latencyBreakDown.EndorseStart.IsZero() || latencyBreakDown.EndorseEnd.IsZero()) {
		latencies.EndorseDuration += int64(latencyBreakDown.EndorseEnd.Sub(latencyBreakDown.EndorseStart))
	}
	if !(latencyBreakDown.CommitStart.IsZero() || latencyBreakDown.CommitEnd.IsZero()) {
		latencies.CommitDuration += int64(latencyBreakDown.CommitEnd.Sub(latencyBreakDown.CommitStart))
	}
	if !(latencyBreakDown.ExecBidlStart.IsZero() || latencyBreakDown.ExecBidlEnd.IsZero()) {
		latencies.ExecBidlDuration += int64(latencyBreakDown.ExecBidlEnd.Sub(latencyBreakDown.ExecBidlStart))
	}
	return latencies
}

func (t *TransactionsProfiling) AddSequenceStart(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].SequenceStart = time.Now()
		} else {
			t.TransactionsProfiling[*transactionID] = &LatencyBreakDown{
				SequenceStart: time.Now(),
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddSequenceEnd(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		t.TransactionsProfiling[*transactionID].SequenceEnd = time.Now()
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddOrderStart(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].OrderStart = time.Now()
		} else {
			t.TransactionsProfiling[*transactionID] = &LatencyBreakDown{
				OrderStart: time.Now(),
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddOrderEndBlock(block *protos.Block) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		for _, tx := range block.Transactions {
			t.TransactionsProfiling[tx.TransactionId].OrderEnd = time.Now()
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddEndorseStart(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].EndorseStart = time.Now()
		} else {
			t.TransactionsProfiling[*transactionID] = &LatencyBreakDown{
				EndorseStart: time.Now(),
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddEndorseStartBlock(block *protos.Block) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		for _, tx := range block.Transactions {
			if _, ok := t.TransactionsProfiling[tx.TransactionId]; ok {
				t.TransactionsProfiling[tx.TransactionId].EndorseStart = time.Now()
			} else {
				t.TransactionsProfiling[tx.TransactionId] = &LatencyBreakDown{
					EndorseStart: time.Now(),
				}
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddEndorseEnd(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].EndorseEnd = time.Now()
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddEndorseEndBlock(block *protos.Block) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		for _, tx := range block.Transactions {
			if _, ok := t.TransactionsProfiling[tx.TransactionId]; ok {
				t.TransactionsProfiling[tx.TransactionId].EndorseEnd = time.Now()
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddExecBidlStart(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].ExecBidlStart = time.Now()
		} else {
			t.TransactionsProfiling[*transactionID] = &LatencyBreakDown{
				ExecBidlStart: time.Now(),
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddExecBidlEnd(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].ExecBidlEnd = time.Now()
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddCommitStart(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].CommitStart = time.Now()
		} else {
			t.TransactionsProfiling[*transactionID] = &LatencyBreakDown{
				CommitStart: time.Now(),
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddCommitStartBlock(block *protos.Block) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		for _, tx := range block.Transactions {
			if _, ok := t.TransactionsProfiling[tx.TransactionId]; ok {
				t.TransactionsProfiling[tx.TransactionId].CommitStart = time.Now()
			} else {
				t.TransactionsProfiling[tx.TransactionId] = &LatencyBreakDown{
					CommitStart: time.Now(),
				}
			}
		}
		t.TransactionsProfilingLock.Unlock()
	}
}

func (t *TransactionsProfiling) AddCommitEnd(transactionID *string) {
	if config.Config.IsTransactionProfilingEnabled {
		t.TransactionsProfilingLock.Lock()
		if _, ok := t.TransactionsProfiling[*transactionID]; ok {
			t.TransactionsProfiling[*transactionID].CommitEnd = time.Now()
		}
		t.TransactionsProfilingLock.Unlock()
	}
}
