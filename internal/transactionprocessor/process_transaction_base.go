package transactionprocessor

import (
	"context"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts"
	"gitlab.lrz.de/orderless/orderlesschain/internal/blockprocessor"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv1"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/keygenerator"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/signer"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transaction"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor/transactiondb"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Processor struct {
	blockProcessor                                  *blockprocessor.Processor
	txJournal                                       *transaction.NodeTransactionJournal
	signer                                          *signer.Signer
	ordererConnectionPool                           map[string]*connpool.Pool
	sequencerConnectionPool                         map[string]*connpool.Pool
	gossipNodesConnectionPool                       map[string]*connpool.Pool
	syncHotStuffNodesConnectionPool                 map[string]*connpool.Pool
	clientProposalSubscribersLock                   *sync.RWMutex
	proposalResponseSubscribers                     map[string]*proposalResponseSubscriber
	clientTransactionSubscribersLock                *sync.RWMutex
	transactionResponseSubscribers                  map[string]*transactionResponseSubscriber
	nodeSubscribersLock                             *sync.RWMutex
	OrderlessChainNodeTransactionResponseSubscriber map[string]*OrderlessChainNodeTransactionResponseSubscriber
	transactionGossipList                           []*protos.Transaction
	transactionGossipListLock                       *sync.Mutex
	sharedShimResources                             *contractinterface.SharedShimResources
	crdtManagerOrderlessChain                       *crdtmanagerv2.Manager
	crdtManagerFabricCRDT                           *crdtmanagerv1.Manager
	inExperimentParticipatingOrderers               []string
	inExperimentParticipatingSequencers             []string
	inExperimentParticipatingNodes                  []string
	inExperimentParticipatingClients                []string
	isThisNodeParticipating                         bool
	gossipNodes                                     []string
	isThisByzantineFailureOrg                       bool
	failureType                                     protos.FailureType
	failureTimer                                    *time.Timer
	failureChannel                                  chan *protos.FailureCommandMode
	failureDoneChannel                              chan bool
	failureRandom                                   *rand.Rand
	clientCountEstimate                             int
	transactionProfiler                             *transaction.TransactionsProfiling
	subscribersLockSyncHotStuff                     *sync.RWMutex
	blockSubscribersSyncHotStuff                    map[string]*blockResponseSubscriberSyncHotStuff
	readyBlockSyncHotStuff                          chan *protos.Block
}

type blockResponseSubscriberSyncHotStuff struct {
	stream   protos.OrdererService_SubscribeBlocksServer
	finished chan<- bool
}

func InitTransactionProcessor() *Processor {
	publicPrivateKey := keygenerator.LoadPublicPrivateKeyFromFile()
	tempProcessor := &Processor{
		clientProposalSubscribersLock:    &sync.RWMutex{},
		proposalResponseSubscribers:      map[string]*proposalResponseSubscriber{},
		clientTransactionSubscribersLock: &sync.RWMutex{},
		signer:                           signer.NewSigner(publicPrivateKey),
		transactionProfiler:              transaction.InitTransactionsProfiling(),
	}
	tempProcessor.setInExperimentParticipatingComponents()
	tempProcessor.setIsThisNodeParticipating()
	if !tempProcessor.isThisNodeParticipating {
		logger.InfoLogger.Println("This node in NOT participating in the experiment")
		return tempProcessor
	}
	tempProcessor.blockProcessor = blockprocessor.InitBlockProcessor()
	tempProcessor.transactionResponseSubscribers = map[string]*transactionResponseSubscriber{}
	tempDBs := map[string]*transactiondb.Operations{}
	for _, contractName := range contracts.GetContractNames() {
		tempDB := transactiondb.NewOperations(contractName)
		tempDBs[contractName] = tempDB
	}
	tempProcessor.sharedShimResources = &contractinterface.SharedShimResources{
		DBConnections: tempDBs,
	}
	tempProcessor.txJournal = transaction.InitTransactionJournal()
	go tempProcessor.txJournal.RunProposalQueueProcessorTicker()
	if config.Config.IsOrderlessChain {
		tempProcessor.OrderlessChainNodeTransactionResponseSubscriber = map[string]*OrderlessChainNodeTransactionResponseSubscriber{}
		tempProcessor.nodeSubscribersLock = &sync.RWMutex{}
		tempProcessor.gossipNodesConnectionPool = connpool.GetNodeConnectionsWatchEventWithoutSelf(tempProcessor.gossipNodes)
		tempProcessor.crdtManagerOrderlessChain = crdtmanagerv2.NewManager(tempDBs)
		tempProcessor.sharedShimResources.CRDTManager = tempProcessor.crdtManagerOrderlessChain
		tempProcessor.transactionGossipList = []*protos.Transaction{}
		tempProcessor.transactionGossipListLock = &sync.Mutex{}
		tempProcessor.failureChannel = make(chan *protos.FailureCommandMode)
		tempProcessor.failureDoneChannel = make(chan bool)
		time.Sleep(10 * time.Second)
		go tempProcessor.txJournal.RunTransactionsQueueProcessorTicker()
		go tempProcessor.runGossipingOrderlessChain()
		go tempProcessor.runTransactionProcessorOrderlessChain()
		go tempProcessor.subscriberForOtherNodeTransactionsOrderlessChain()
		go tempProcessor.runProposalQueueProcessingOrderlessChain()
		go tempProcessor.runTransactionQueueProcessingOrderlessChain()
		go tempProcessor.runFailureExecutionMonitor()
	} else if config.Config.IsFabric || config.Config.IsFabricCRDT {
		tempProcessor.ordererConnectionPool = connpool.GetOrdererConnections(tempProcessor.inExperimentParticipatingOrderers)
		if config.Config.IsFabricCRDT {
			tempProcessor.crdtManagerFabricCRDT = crdtmanagerv1.NewManager()
			tempProcessor.sharedShimResources.CRDTManagerFabricCRDT = tempProcessor.crdtManagerFabricCRDT
			go tempProcessor.runProposalQueueProcessingFabricCRDT()
		} else {
			go tempProcessor.runProposalQueueProcessingFabric()
		}
		go tempProcessor.runTransactionProcessorFabricAndFabricCRDT()
		go tempProcessor.subscriberForBlockEventsFabricAndFabricCRDT()
		go tempProcessor.txJournal.RunPopFirstBlock()
	} else if config.Config.IsBIDL {
		tempProcessor.ordererConnectionPool = connpool.GetOrdererConnections(tempProcessor.inExperimentParticipatingOrderers)
		tempProcessor.sequencerConnectionPool = connpool.GetSequencerConnections(tempProcessor.inExperimentParticipatingSequencers)
		go tempProcessor.runProposalQueueProcessingBIDL()
		go tempProcessor.runTransactionProcessorBIDL()
		go tempProcessor.subscriberForBlockEventsBILD()
		go tempProcessor.subscriberForSequencedTransactions()
		go tempProcessor.txJournal.RunPopFirstBlock()

	} else if config.Config.IsSyncHotStuff {
		tempProcessor.subscribersLockSyncHotStuff = &sync.RWMutex{}
		tempProcessor.blockSubscribersSyncHotStuff = map[string]*blockResponseSubscriberSyncHotStuff{}
		tempProcessor.syncHotStuffNodesConnectionPool = connpool.GetNodeConnectionsWatchEventWithSelf(tempProcessor.inExperimentParticipatingNodes)
		tempProcessor.ordererConnectionPool = connpool.GetOrdererConnections(tempProcessor.inExperimentParticipatingOrderers)
		tempProcessor.readyBlockSyncHotStuff = make(chan *protos.Block)
		go tempProcessor.subscriberForBlockEventsSyncHotStuff()
		go tempProcessor.subscribeForOtherNodeBlockSyncHotStuff()
		go tempProcessor.runProcessBlockSyncHotStuff()
	} else {
		logger.FatalLogger.Fatalln("target system not set")
	}
	return tempProcessor
}

type proposalResponseSubscriber struct {
	stream   protos.TransactionService_SubscribeProposalResponseServer
	finished chan<- bool
}

type transactionResponseSubscriber struct {
	stream   protos.TransactionService_SubscribeTransactionResponseServer
	finished chan<- bool
}

func (p *Processor) setInExperimentParticipatingComponents() {
	for i := 0; i < config.Config.TotalOrdererCount; i++ {
		p.inExperimentParticipatingOrderers = append(p.inExperimentParticipatingOrderers, "orderer"+strconv.Itoa(i))
	}
	for i := 0; i < config.Config.TotalSequencerCount; i++ {
		p.inExperimentParticipatingSequencers = append(p.inExperimentParticipatingSequencers, "sequencer"+strconv.Itoa(i))
	}
	for i := 0; i < config.Config.TotalNodeCount; i++ {
		p.inExperimentParticipatingNodes = append(p.inExperimentParticipatingNodes, "node"+strconv.Itoa(i))
	}
	for i := 0; i < config.Config.TotalClientCount; i++ {
		p.inExperimentParticipatingClients = append(p.inExperimentParticipatingClients, "client"+strconv.Itoa(i))
	}
	currentNodeId := connpool.GetComponentPseudoName()
	currentNodeId = strings.ReplaceAll(currentNodeId, "node", "")
	nodeCount := len(p.inExperimentParticipatingNodes)
	currentNodeIdInt, err := strconv.Atoi(currentNodeId)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	currentNode := currentNodeIdInt
	currentNode += config.Config.EndorsementPolicy

	for i := 0; i < config.Config.GossipNodeCount; i++ {
		currentNode++
		nodeId := currentNode % nodeCount
		if nodeId == currentNodeIdInt {
			currentNode++
			nodeId = currentNode % nodeCount
		}
		p.gossipNodes = append(p.gossipNodes, "node"+strconv.Itoa(nodeId))
	}
}

func (p *Processor) setIsThisNodeParticipating() {
	currentNodeId := connpool.GetComponentPseudoName()
	for _, node := range p.inExperimentParticipatingNodes {
		if currentNodeId == node {
			p.isThisNodeParticipating = true
		}
	}
}

func (p *Processor) LoadOtherNodesPublicKeys() {
	allConnectionNodes := connpool.GetAllNodesConnections()
	for name := range allConnectionNodes {
		conn, err := allConnectionNodes[name].Get(context.Background())
		if conn == nil || err != nil {
			connpool.SleepAndReconnect()
			p.LoadOtherNodesPublicKeys()
			return
		}
		client := protos.NewTransactionServiceClient(conn.ClientConn)
		publicKey, err := client.GetPublicKey(context.Background(), &protos.Empty{})
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(errCon)
		}
		if err != nil {
			connpool.SleepAndReconnect()
			p.LoadOtherNodesPublicKeys()
			return
		}
		p.signer.AddPublicKey(publicKey.NodeId, publicKey.PublicKey)
	}
}

func (p *Processor) makeFailedProposal(proposalId string) *protos.ProposalResponse {
	return &protos.ProposalResponse{ProposalId: proposalId, Status: protos.ProposalResponse_FAIL, NodeId: config.Config.UUID}
}

func (p *Processor) executeContract(pr *protos.ProposalRequest) (*protos.ProposalResponse, error) {
	if contractInterface, err := contracts.GetContract(pr.ContractName); err == nil {
		contractCode := contractInterface.(contractinterface.ContractInterface)
		proposalResponse, invokeErr := contractCode.Invoke(contract.NewShim(pr, p.sharedShimResources, pr.ContractName), pr)
		if invokeErr != nil {
			logger.ErrorLogger.Println(invokeErr)
			return p.makeFailedProposal(pr.ProposalId), invokeErr
		}
		if proposalResponse != nil {
			return proposalResponse, nil
		}
	}
	return p.makeFailedProposal(pr.ProposalId), errors.New("proposal execution failed")
}

func (p *Processor) makeFailedTransactionResponse(transactionID string, status protos.TransactionStatus, blockHeader []byte) *protos.TransactionResponse {
	return &protos.TransactionResponse{
		BlockHeader:   blockHeader,
		TransactionId: transactionID,
		Status:        status,
		NodeId:        config.Config.UUID,
	}
}

func (p *Processor) makeSuccessTransactionResponse(transactionID string, blockHeader []byte) *protos.TransactionResponse {
	return &protos.TransactionResponse{
		BlockHeader:   blockHeader,
		TransactionId: transactionID,
		Status:        protos.TransactionStatus_SUCCEEDED,
		NodeId:        config.Config.UUID,
	}
}

func (p *Processor) ProposalResponseSubscription(subscription *protos.ProposalResponseEventSubscription,
	stream protos.TransactionService_SubscribeProposalResponseServer) error {
	finished := make(chan bool)
	p.clientProposalSubscribersLock.Lock()
	p.proposalResponseSubscribers[subscription.ComponentId] = &proposalResponseSubscriber{
		stream:   stream,
		finished: finished,
	}
	p.clientProposalSubscribersLock.Unlock()
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

func (p *Processor) sendProposalResponseToSubscriber(clientID string, response *protos.ProposalResponse) {
	p.clientProposalSubscribersLock.RLock()
	streamer, ok := p.proposalResponseSubscribers[clientID]
	p.clientProposalSubscribersLock.RUnlock()
	if !ok {
		return
	}
	if err := streamer.stream.Send(response); err != nil {
		streamer.finished <- true
		logger.ErrorLogger.Println("Could not send the proposal response to the client " + clientID)
		p.clientProposalSubscribersLock.Lock()
		delete(p.proposalResponseSubscribers, clientID)
		p.clientProposalSubscribersLock.Unlock()
	}
	if response.Status == protos.ProposalResponse_SUCCESS {
		p.transactionProfiler.AddEndorseEnd(&response.ProposalId)
	}
}

func (p *Processor) TransactionResponseSubscription(subscription *protos.TransactionResponseEventSubscription,
	stream protos.TransactionService_SubscribeTransactionResponseServer) error {
	finished := make(chan bool)
	p.clientTransactionSubscribersLock.Lock()
	p.transactionResponseSubscribers[subscription.ComponentId] = &transactionResponseSubscriber{
		stream:   stream,
		finished: finished,
	}
	p.clientTransactionSubscribersLock.Unlock()
	p.signer.AddPublicKey(subscription.ComponentId, subscription.PublicKey)
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

func (p *Processor) AddClientPublicKey(subscription *protos.TransactionResponseEventSubscription) error {
	p.signer.AddPublicKey(subscription.ComponentId, subscription.PublicKey)
	return nil
}

func (p *Processor) sendTransactionResponseToSubscriber(clientID string, response *protos.TransactionResponse) {
	p.clientTransactionSubscribersLock.RLock()
	streamer, ok := p.transactionResponseSubscribers[clientID]
	p.clientTransactionSubscribersLock.RUnlock()
	if !ok {
		return
	}
	if err := streamer.stream.Send(response); err != nil {
		streamer.finished <- true
		logger.ErrorLogger.Println("Could not send the transaction response to the client " + clientID)
		p.clientTransactionSubscribersLock.Lock()
		delete(p.transactionResponseSubscribers, clientID)
		p.clientTransactionSubscribersLock.Unlock()
	}
	if response.Status == protos.TransactionStatus_SUCCEEDED {
		p.transactionProfiler.AddCommitEnd(&response.TransactionId)
	}
}

func (p *Processor) runFailureExecutionMonitor() {
	for {
		failureCommand := <-p.failureChannel
		logger.InfoLogger.Println("Running Failure: ", failureCommand)
		p.failureType = failureCommand.FailureType
		p.failureRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
		p.failureTimer = time.NewTimer(time.Duration(failureCommand.FailureDurationS) * time.Second)
		p.isThisByzantineFailureOrg = true
		select {
		case <-p.failureDoneChannel:
			p.failureTimer.Stop()
		case <-p.failureTimer.C:
			logger.InfoLogger.Println("Failure over.")
		}
		p.isThisByzantineFailureOrg = false
	}
}

func (p *Processor) SetFailureCommand(command *protos.FailureCommandMode) {
	if p.isThisByzantineFailureOrg {
		p.failureDoneChannel <- true
	}
	p.failureChannel <- command
}

func (p *Processor) ShouldFailByzantineNetwork() bool {
	if p.isThisByzantineFailureOrg {
		if p.failureType == protos.FailureType_RANODM {
			if p.failureRandom.Intn(2) == 0 {
				return true
			}
			return false
		}
		return p.failureType == protos.FailureType_NOTRESPONDING || p.failureType == protos.FailureType_CRASHED
	}
	return false
}

func (p *Processor) ShouldFailByzantineTampered() bool {
	if p.isThisByzantineFailureOrg {
		if p.failureType == protos.FailureType_RANODM {
			return true
		}
		return p.failureType == protos.FailureType_TAMPERED
	}
	return false
}

func (p *Processor) SendTransactionProfiling(stream protos.TransactionService_GetTransactionProfilingResultServer) error {
	p.transactionProfiler.TransactionsProfilingLock.Lock()
	defer p.transactionProfiler.TransactionsProfilingLock.Unlock()
	var err error
	for transactionId, txProfile := range p.transactionProfiler.TransactionsProfiling {
		transactionProfile := p.transactionProfiler.PrepareBreakdownToBeSent(transactionId, txProfile)
		if transactionProfile != nil {
			if err = stream.Send(transactionProfile); err != nil {
				break
			}
		}
	}
	return err
}
