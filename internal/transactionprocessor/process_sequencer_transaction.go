package transactionprocessor

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transaction"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"reflect"
	"strconv"
	"sync"
)

type transactionsForOrderingSubscriber struct {
	stream   protos.SequencerService_SubscribeTransactionsForOrderingServer
	finished chan<- bool
}

type transactionsForCommittingSubscriber struct {
	stream   protos.SequencerService_SubscribeTransactionsForProcessingServer
	finished chan<- bool
}

type SequencerProcessor struct {
	sequencer       int32
	transactionChan chan *protos.ProposalRequest

	subscribersLock                      *sync.RWMutex
	transactionsForOrderingSubscribers   map[string]*transactionsForOrderingSubscriber
	transactionsForCommittingSubscribers map[string]*transactionsForCommittingSubscriber

	inExperimentParticipatingOrderers   []string
	inExperimentParticipatingSequencers []string
	inExperimentParticipatingNodes      []string
	inExperimentParticipatingClients    []string
	isThisSequencerParticipating        bool

	transactionProfiler *transaction.TransactionsProfiling
}

func InitTransactionSequencerProcessor() *SequencerProcessor {
	tempProcessor := &SequencerProcessor{
		transactionChan:                      make(chan *protos.ProposalRequest, 10000),
		subscribersLock:                      &sync.RWMutex{},
		transactionsForOrderingSubscribers:   map[string]*transactionsForOrderingSubscriber{},
		transactionsForCommittingSubscribers: map[string]*transactionsForCommittingSubscriber{},
		transactionProfiler:                  transaction.InitTransactionsProfiling(),
	}
	tempProcessor.setInExperimentParticipatingComponents()
	tempProcessor.setIsThisSequencerParticipating()
	if !tempProcessor.isThisSequencerParticipating {
		logger.InfoLogger.Println("This sequencer in NOT participating in the experiment")
		return tempProcessor
	}
	go tempProcessor.runTransactionQueueProcessing()
	return tempProcessor
}

func (s *SequencerProcessor) ProcessTransactionStream(tx *protos.ProposalRequest) {
	s.transactionProfiler.AddSequenceStart(&tx.ProposalId)
	s.transactionChan <- tx
}

func (s *SequencerProcessor) runTransactionQueueProcessing() {
	for {
		tx := <-s.transactionChan
		s.sequencer++
		tx.TransactionSequence = s.sequencer
		go s.sendDequeuedTransactions(tx)
	}
}

func (s *SequencerProcessor) sendDequeuedTransactions(tx *protos.ProposalRequest) {
	s.subscribersLock.RLock()
	orderers := reflect.ValueOf(s.transactionsForOrderingSubscribers).MapKeys()
	nodes := reflect.ValueOf(s.transactionsForCommittingSubscribers).MapKeys()
	s.subscribersLock.RUnlock()
	for _, orderer := range orderers {
		ordererId := orderer.String()

		s.subscribersLock.RLock()
		streamer, ok := s.transactionsForOrderingSubscribers[ordererId]
		s.subscribersLock.RUnlock()

		if !ok {
			logger.ErrorLogger.Println("Orderer was not found in the subscribers streams.", ordererId)
			return
		}
		if err := streamer.stream.Send(tx); err != nil {
			streamer.finished <- true
			logger.ErrorLogger.Println("Could not send the response to the orderer " + ordererId)
			s.subscribersLock.Lock()
			delete(s.transactionsForOrderingSubscribers, ordererId)
			s.subscribersLock.Unlock()
		}
	}

	for _, node := range nodes {
		nodeId := node.String()

		s.subscribersLock.RLock()
		streamer, ok := s.transactionsForCommittingSubscribers[nodeId]
		s.subscribersLock.RUnlock()

		if !ok {
			logger.ErrorLogger.Println("Node was not found in the subscribers streams.", nodeId)
			return
		}
		if err := streamer.stream.Send(tx); err != nil {
			streamer.finished <- true
			logger.ErrorLogger.Println("Could not send the response to the node " + nodeId)
			s.subscribersLock.Lock()
			delete(s.transactionsForCommittingSubscribers, nodeId)
			s.subscribersLock.Unlock()
		}
	}
	s.transactionProfiler.AddSequenceEnd(&tx.ProposalId)
}

func (s *SequencerProcessor) SubscribeForOrdering(subscriptionEvent *protos.SequencedTransactionForOrdering,
	stream protos.SequencerService_SubscribeTransactionsForOrderingServer) error {
	finished := make(chan bool)
	s.subscribersLock.Lock()
	s.transactionsForOrderingSubscribers[subscriptionEvent.NodeId] = &transactionsForOrderingSubscriber{
		stream:   stream,
		finished: finished,
	}
	s.subscribersLock.Unlock()
	cntx := stream.Context()
	for {
		select {
		case <-finished:
			return nil
		case <-cntx.Done():
			return nil
		}
	}
}

func (s *SequencerProcessor) SubscribeForCommitting(subscriptionEvent *protos.SequencedTransactionForCommitting,
	stream protos.SequencerService_SubscribeTransactionsForProcessingServer) error {
	finished := make(chan bool)
	s.subscribersLock.Lock()
	s.transactionsForCommittingSubscribers[subscriptionEvent.NodeId] = &transactionsForCommittingSubscriber{
		stream:   stream,
		finished: finished,
	}
	s.subscribersLock.Unlock()
	cntx := stream.Context()
	for {
		select {
		case <-finished:
			return nil
		case <-cntx.Done():
			return nil
		}
	}
}

func (s *SequencerProcessor) setInExperimentParticipatingComponents() {
	for i := 0; i < config.Config.TotalOrdererCount; i++ {
		s.inExperimentParticipatingOrderers = append(s.inExperimentParticipatingOrderers, "orderer"+strconv.Itoa(i))
	}
	for i := 0; i < config.Config.TotalSequencerCount; i++ {
		s.inExperimentParticipatingSequencers = append(s.inExperimentParticipatingSequencers, "sequencer"+strconv.Itoa(i))
	}
	for i := 0; i < config.Config.TotalNodeCount; i++ {
		s.inExperimentParticipatingNodes = append(s.inExperimentParticipatingNodes, "node"+strconv.Itoa(i))
	}
	for i := 0; i < config.Config.TotalClientCount; i++ {
		s.inExperimentParticipatingClients = append(s.inExperimentParticipatingClients, "client"+strconv.Itoa(i))
	}
}

func (s *SequencerProcessor) setIsThisSequencerParticipating() {
	currentSequencerId := connpool.GetComponentPseudoName()
	for _, sequencer := range s.inExperimentParticipatingSequencers {
		if currentSequencerId == sequencer {
			s.isThisSequencerParticipating = true
		}
	}
}

func (s *SequencerProcessor) SendTransactionProfiling(stream protos.SequencerService_GetTransactionProfilingResultServer) error {
	s.transactionProfiler.TransactionsProfilingLock.Lock()
	defer s.transactionProfiler.TransactionsProfilingLock.Unlock()
	var err error
	for transactionId, txProfile := range s.transactionProfiler.TransactionsProfiling {
		transactionProfile := s.transactionProfiler.PrepareBreakdownToBeSent(transactionId, txProfile)
		if transactionProfile != nil {
			if err = stream.Send(transactionProfile); err != nil {
				break
			}
		}
	}
	return err
}
