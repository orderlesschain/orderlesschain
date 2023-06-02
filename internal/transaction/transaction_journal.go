package transaction

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sort"
	"sync"
	"time"
)

type DequeuedProposals struct {
	DequeuedProposals []*protos.ProposalRequest
}

type DequeuedTransactions struct {
	DequeuedTransactions []*protos.Transaction
}

type QueuedBlockSynHotStuff struct {
	block      *protos.Block
	votedNodes map[string]bool
	processed  bool
}

type NodeTransactionJournal struct {
	journal                        map[string]bool
	journalNodes                   map[string]bool
	journalBIDL                    map[string]*protos.Transaction
	journalLock                    *sync.Mutex
	journalNodesLock               *sync.Mutex
	proposalsQueue                 []*protos.ProposalRequest
	transactionsQueue              []*protos.Transaction
	blockQueue                     []*protos.Block
	DequeuedProposalsChan          chan *DequeuedProposals
	DequeuedTransactionsChan       chan *DequeuedTransactions
	DequeuedOrdererBlockChan       chan *protos.Block
	RequestBlockChan               chan bool
	proposalsQueueLock             *sync.Mutex
	transactionsQueueLock          *sync.Mutex
	blockQueueLock                 *sync.Mutex
	tickerDuration                 time.Duration
	proposalBatchSizePertTicker    int
	transactionBatchSizePertTicker int
	blocksSyncHotStuff             map[string]*QueuedBlockSynHotStuff
}

func InitTransactionJournal() *NodeTransactionJournal {
	tempJournal := &NodeTransactionJournal{
		journal:                  map[string]bool{},
		journalNodes:             map[string]bool{},
		journalBIDL:              map[string]*protos.Transaction{},
		journalLock:              &sync.Mutex{},
		journalNodesLock:         &sync.Mutex{},
		proposalsQueue:           []*protos.ProposalRequest{},
		transactionsQueue:        []*protos.Transaction{},
		blockQueue:               []*protos.Block{},
		DequeuedProposalsChan:    make(chan *DequeuedProposals),
		DequeuedTransactionsChan: make(chan *DequeuedTransactions),
		DequeuedOrdererBlockChan: make(chan *protos.Block),
		RequestBlockChan:         make(chan bool),
		proposalsQueueLock:       &sync.Mutex{},
		transactionsQueueLock:    &sync.Mutex{},
		blockQueueLock:           &sync.Mutex{},
		blocksSyncHotStuff:       map[string]*QueuedBlockSynHotStuff{},
	}
	tickerDurationMS := config.Config.QueueTickerDurationMS
	tempJournal.tickerDuration = time.Duration(tickerDurationMS) * time.Millisecond
	tempJournal.proposalBatchSizePertTicker = config.Config.ProposalQueueConsumptionRateTPS / (1000 / tickerDurationMS)
	tempJournal.transactionBatchSizePertTicker = config.Config.TransactionQueueConsumptionRateTPS / (1000 / tickerDurationMS)

	return tempJournal
}

func (tj *NodeTransactionJournal) AddTransactionToJournalIfNotExist(transaction *protos.Transaction) bool {
	tj.journalLock.Lock()
	if _, ok := tj.journal[transaction.TransactionId]; !ok {
		tj.journal[transaction.TransactionId] = true
		tj.journalLock.Unlock()
		return true
	}
	tj.journalLock.Unlock()
	return false
}

func (tj *NodeTransactionJournal) IsTransactionInJournal(transaction *protos.Transaction) bool {
	tj.journalLock.Lock()
	_, ok := tj.journal[transaction.TransactionId]
	tj.journalLock.Unlock()
	return ok
}

func (tj *NodeTransactionJournal) AddTransactionToJournalNodeIfNotExist(transaction *protos.Transaction) bool {
	tj.journalNodesLock.Lock()
	if _, ok := tj.journalNodes[transaction.TransactionId]; !ok {
		tj.journalNodes[transaction.TransactionId] = true
		tj.journalNodesLock.Unlock()
		return true
	}
	tj.journalNodesLock.Unlock()
	return false
}

func (tj *NodeTransactionJournal) AddBILDTransactionFromSequencerAndCheckIfCanBeCommitted(transaction *protos.Transaction) *protos.Transaction {
	var finalTransaction *protos.Transaction
	tj.journalLock.Lock()
	if _, ok := tj.journalBIDL[transaction.TransactionId]; !ok {
		tj.journalBIDL[transaction.TransactionId] = transaction
		tj.journalLock.Unlock()
		return nil
	} else {
		finalTransaction = transaction
		delete(tj.journalBIDL, transaction.TransactionId)
	}
	tj.journalLock.Unlock()
	return finalTransaction
}

func (tj *NodeTransactionJournal) AddBILDTransactionFromOrdererAndCheckIfCanBeCommitted(transaction *protos.Transaction) *protos.Transaction {
	var finalTransaction *protos.Transaction
	tj.journalLock.Lock()
	if existingTransaction, ok := tj.journalBIDL[transaction.TransactionId]; !ok {
		tj.journalBIDL[transaction.TransactionId] = transaction
		tj.journalLock.Unlock()
		return nil
	} else {
		finalTransaction = existingTransaction
		delete(tj.journalBIDL, transaction.TransactionId)
	}
	tj.journalLock.Unlock()
	return finalTransaction
}

func (tj *NodeTransactionJournal) AddProposalToQueue(proposal *protos.ProposalRequest) {
	tj.proposalsQueueLock.Lock()
	tj.proposalsQueue = append(tj.proposalsQueue, proposal)
	tj.proposalsQueueLock.Unlock()
}

func (tj *NodeTransactionJournal) AddTransactionToQueue(transaction *protos.Transaction) {
	tj.transactionsQueueLock.Lock()
	tj.transactionsQueue = append(tj.transactionsQueue, transaction)
	tj.transactionsQueueLock.Unlock()
}

func (tj *NodeTransactionJournal) AddBlockToQueue(block *protos.Block) {
	tj.blockQueueLock.Lock()
	tj.blockQueue = append(tj.blockQueue, block)
	tj.blockQueueLock.Unlock()
}

func (tj *NodeTransactionJournal) AddBlockToQueueSyncHotStuff(block *protos.Block) {
	tj.blockQueueLock.Lock()
	if _, ok := tj.blocksSyncHotStuff[block.BlockId]; !ok {
		tj.blocksSyncHotStuff[block.BlockId] = &QueuedBlockSynHotStuff{
			block:      block,
			votedNodes: map[string]bool{},
		}
	} else {
		tj.blocksSyncHotStuff[block.BlockId].block = block
	}
	tj.blockQueueLock.Unlock()
}

func (tj *NodeTransactionJournal) AddNodeVoteSyncHotStuff(blockId string, nodeId string, quorum int) *protos.Block {
	var existingBlock *QueuedBlockSynHotStuff
	var confirmedBlock *protos.Block
	var ok bool
	tj.blockQueueLock.Lock()
	if existingBlock, ok = tj.blocksSyncHotStuff[blockId]; ok {
		existingBlock.votedNodes[nodeId] = true
	} else {
		tj.blocksSyncHotStuff[blockId] = &QueuedBlockSynHotStuff{
			votedNodes: map[string]bool{nodeId: true},
		}
		existingBlock = tj.blocksSyncHotStuff[blockId]
	}
	if !existingBlock.processed && len(existingBlock.votedNodes) >= quorum {
		existingBlock.processed = true
		confirmedBlock = existingBlock.block
	}
	tj.blockQueueLock.Unlock()
	return confirmedBlock
}

func (tj *NodeTransactionJournal) RunProposalQueueProcessorTicker() {
	ticker := time.NewTicker(tj.tickerDuration)
	for range ticker.C {
		tj.proposalsQueueLock.Lock()
		toDequeueLength := len(tj.proposalsQueue)
		if toDequeueLength == 0 {
			tj.proposalsQueueLock.Unlock()
			continue
		}
		if toDequeueLength > tj.proposalBatchSizePertTicker {
			toDequeueLength = tj.proposalBatchSizePertTicker
		}
		tempDequeuedProposals := &DequeuedProposals{
			DequeuedProposals: make([]*protos.ProposalRequest, 0, toDequeueLength),
		}
		for i := 0; i < toDequeueLength; i++ {
			tempDequeuedProposals.DequeuedProposals = append(tempDequeuedProposals.DequeuedProposals, tj.proposalsQueue[i])
			tj.proposalsQueue[i] = nil
		}
		tj.proposalsQueue = tj.proposalsQueue[toDequeueLength:]
		tj.proposalsQueueLock.Unlock()
		tj.DequeuedProposalsChan <- tempDequeuedProposals
	}
}

func (tj *NodeTransactionJournal) RunTransactionsQueueProcessorTicker() {
	ticker := time.NewTicker(tj.tickerDuration)
	for range ticker.C {
		tj.transactionsQueueLock.Lock()
		toDequeueLength := len(tj.transactionsQueue)
		if toDequeueLength == 0 {
			tj.transactionsQueueLock.Unlock()
			continue
		}
		if toDequeueLength > tj.transactionBatchSizePertTicker {
			toDequeueLength = tj.transactionBatchSizePertTicker
		}
		tempDequeuedTransactions := &DequeuedTransactions{
			DequeuedTransactions: make([]*protos.Transaction, 0, toDequeueLength),
		}
		for i := 0; i < toDequeueLength; i++ {
			tempDequeuedTransactions.DequeuedTransactions = append(tempDequeuedTransactions.DequeuedTransactions, tj.transactionsQueue[i])
			tj.transactionsQueue[i] = nil
		}
		tj.transactionsQueue = tj.transactionsQueue[toDequeueLength:]
		tj.transactionsQueueLock.Unlock()
		tj.DequeuedTransactionsChan <- tempDequeuedTransactions
	}
}

func (tj *NodeTransactionJournal) RunPopFirstBlock() {
	for {
		<-tj.RequestBlockChan
		var block *protos.Block
		tj.blockQueueLock.Lock()
		if len(tj.blockQueue) > 0 {
			block = tj.blockQueue[0]
			tj.blockQueue[0] = nil
			tj.blockQueue = tj.blockQueue[1:]
		}
		tj.blockQueueLock.Unlock()
		tj.DequeuedOrdererBlockChan <- block
	}
}

func (tj *NodeTransactionJournal) MakeTransactionFromProposalForBILDAnSyncHotStuff(proposal *protos.ProposalRequest,
	proposalResponse *protos.ProposalResponse) (*protos.Transaction, error) {

	tempTransaction := &protos.Transaction{
		TargetSystem:   protos.TargetSystem_BIDL,
		TransactionId:  proposalResponse.ProposalId,
		ClientId:       proposal.ClientId,
		ContractName:   proposal.ContractName,
		NodeSignatures: map[string][]byte{},
	}

	tempTransaction.ReadWriteSet = proposalResponse.ReadWriteSet
	sort.Slice(tempTransaction.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	return tempTransaction, nil
}
