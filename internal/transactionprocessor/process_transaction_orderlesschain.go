package transactionprocessor

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlesschain/internal/blockprocessor"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/profiling"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transaction"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor/transactiondb"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"time"
)

var byzantineTemperedMessage = []byte("tempered")

type OrderlessChainNodeTransactionResponseSubscriber struct {
	stream   protos.TransactionService_SubscribeNodeTransactionsServer
	finished chan<- bool
}

func (p *Processor) signProposalResponseOrderlessChain(proposalResponse *protos.ProposalResponse) (*protos.ProposalResponse, error) {
	sort.Slice(proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return proposalResponse.ReadWriteSet.ReadKeys.ReadKeys[i].Key < proposalResponse.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	marshalledReadWriteSet, err := proto.Marshal(proposalResponse.ReadWriteSet)
	if err != nil {
		return p.makeFailedProposal(proposalResponse.ProposalId), nil
	}
	if p.ShouldFailByzantineTampered() {
		marshalledReadWriteSet = append(marshalledReadWriteSet, byzantineTemperedMessage...)
	}
	proposalResponse.NodeSignature = p.signer.Sign(marshalledReadWriteSet)
	return proposalResponse, nil
}

func (p *Processor) preProcessValidateReadWriteSetOrderlessChain(tx *protos.Transaction) error {
	if p.ShouldFailByzantineTampered() {
		return nil
	}
	sort.Slice(tx.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tx.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tx.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(tx.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return tx.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < tx.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	txDigest, err := proto.Marshal(tx.ReadWriteSet)
	if err != nil {
		return err
	}
	passedSignature := int32(0)
	for node, nodeSign := range tx.NodeSignatures {
		err = p.signer.Verify(node, txDigest, nodeSign)
		if err == nil {
			passedSignature++
		}
	}
	if passedSignature < tx.EndorsementPolicy {
		return errors.New("failed signature validation")
	}
	err = p.signer.Verify(tx.ClientId, txDigest, tx.ClientSignature)
	if err != nil {
		return err
	}
	return nil
}

func (p *Processor) ProcessProposalOrderlessChainStream(proposal *protos.ProposalRequest) {
	p.transactionProfiler.AddEndorseStart(&proposal.ProposalId)
	p.txJournal.AddProposalToQueue(proposal)
}

func (p *Processor) runProposalQueueProcessingOrderlessChain() {
	for {
		proposals := <-p.txJournal.DequeuedProposalsChan
		go p.processDequeuedProposalsOrderlessChain(proposals)
	}
}

func (p *Processor) processDequeuedProposalsOrderlessChain(proposals *transaction.DequeuedProposals) {
	for _, proposal := range proposals.DequeuedProposals {
		go p.processProposalOrderlessChain(proposal)
	}
}

func (p *Processor) processProposalOrderlessChain(proposal *protos.ProposalRequest) {
	response, err := p.executeContract(proposal)
	if err != nil {
		p.sendProposalResponseToSubscriber(proposal.ClientId, response)
		return
	}
	response, err = p.signProposalResponseOrderlessChain(response)
	if err != nil {
		p.sendProposalResponseToSubscriber(proposal.ClientId, response)
		return
	}
	p.sendProposalResponseToSubscriber(proposal.ClientId, response)
}

func (p *Processor) ProcessTransactionOrderlessChainStream(tx *protos.Transaction) {
	p.transactionProfiler.AddCommitStart(&tx.TransactionId)
	p.txJournal.AddTransactionToQueue(tx)
}

func (p *Processor) runTransactionQueueProcessingOrderlessChain() {
	for {
		transactions := <-p.txJournal.DequeuedTransactionsChan
		go p.processDequeuedTransactionsOrderlessChain(transactions)
	}
}

func (p *Processor) processDequeuedTransactionsOrderlessChain(transactions *transaction.DequeuedTransactions) {
	for _, tx := range transactions.DequeuedTransactions {
		go p.processTransactionOrderlessChain(tx)
	}
}

func (p *Processor) processTransactionOrderlessChain(tx *protos.Transaction) {
	if !tx.FromNode {
		if err := p.preProcessValidateReadWriteSetOrderlessChain(tx); err != nil {
			tx.Status = protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION
		}
	}
	shouldBeAdded := p.txJournal.AddTransactionToJournalIfNotExist(tx)
	if shouldBeAdded {
		p.blockProcessor.TransactionChan <- tx
	} else {
		if !tx.FromNode {
			p.finalizeAlreadySubmittedTransactionFromClientOrderlessChain(tx)
		}
	}
}

func (p *Processor) processTransactionFromOtherNodesOrderlessChain(txs []*protos.Transaction) {
	for _, tx := range txs {
		shouldBeAdded := p.txJournal.AddTransactionToJournalNodeIfNotExist(tx)
		if shouldBeAdded {
			alreadyInJournal := p.txJournal.IsTransactionInJournal(tx)
			if !alreadyInJournal {
				tx.FromNode = true
				p.txJournal.AddTransactionToQueue(tx)
			}
		}
	}
}

func (p *Processor) runTransactionProcessorOrderlessChain() {
	for {
		block := <-p.blockProcessor.MinedBlock
		go p.processBlockOrderlessChain(block)
	}
}

func (p *Processor) processBlockOrderlessChain(block *blockprocessor.MinedBlock) {
	for _, tx := range block.Transactions {
		go p.processTransactionInBlockOrderlessChain(tx, block)
	}
}

func (p *Processor) processTransactionInBlockOrderlessChain(tx *protos.Transaction, block *blockprocessor.MinedBlock) {
	if tx.Status == protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
		p.finalizeTransactionResponseOrderlessChain(tx.ClientId,
			p.makeFailedTransactionResponse(tx.TransactionId, tx.Status, block.Block.ThisBlockHash), tx)
	} else {
		response, err := p.commitTransactionOrderlessChain(tx)
		if err == nil {
			response.BlockHeader = block.Block.ThisBlockHash
			p.finalizeTransactionResponseOrderlessChain(tx.ClientId, response, tx)
		} else {
			p.finalizeTransactionResponseOrderlessChain(tx.ClientId,
				p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_DATABASE, block.Block.ThisBlockHash), tx)
		}
	}
}

func (p *Processor) commitTransactionOrderlessChain(tx *protos.Transaction) (*protos.TransactionResponse, error) {
	if err := p.addTransactionToDatabaseOrderlessChain(tx); err != nil {
		logger.InfoLogger.Println(err)
		return nil, err
	}
	return p.makeSuccessTransactionResponse(tx.TransactionId, []byte{}), nil
}

func (p *Processor) addTransactionToDatabaseOrderlessChain(tx *protos.Transaction) error {
	dbOp := p.sharedShimResources.DBConnections[tx.ContractName]
	var err error
	for _, keyValue := range tx.ReadWriteSet.WriteKeyValues.WriteKeyValues {
		err = p.addKeyValueToDatabaseOrderlessChain(keyValue, tx, dbOp)
	}
	return err
}

func (p *Processor) addKeyValueToDatabaseWithWaitGroupOrderlessChain(keyValue *protos.WriteKeyValue, tx *protos.Transaction, dbOp *transactiondb.Operations, errWg *error, wg *sync.WaitGroup) {
	err := p.addKeyValueToDatabaseOrderlessChain(keyValue, tx, dbOp)
	if err != nil {
		errWg = &err
	}
	wg.Done()
}

func (p *Processor) addKeyValueToDatabaseOrderlessChain(keyValue *protos.WriteKeyValue, tx *protos.Transaction, dbOp *transactiondb.Operations) error {
	switch keyValue.WriteType {
	case protos.WriteKeyValue_CRDTOPERATIONSLIST_WARM:
		operationsList := &protos.CRDTOperationsList{}
		if err := proto.Unmarshal(keyValue.Value, operationsList); err != nil {
			logger.ErrorLogger.Println(err)
			return err
		}
		p.crdtManagerOrderlessChain.ApplyOperationsWarm(crdtmanagerv2.NewCRDTOperationsPut(tx.ContractName, operationsList))
		if err := dbOp.PutKeyValueNoVersion(keyValue.Key+"-"+strconv.FormatInt(time.Now().UnixNano(), 10),
			keyValue.Value); err != nil {
			logger.ErrorLogger.Println(err)
			return err
		}
	case protos.WriteKeyValue_CRDTOPERATIONSLIST_COLD:
		if err := dbOp.PutKeyValueNoVersion(keyValue.Key+"-"+strconv.FormatInt(time.Now().UnixNano(), 10),
			keyValue.Value); err != nil {
			logger.ErrorLogger.Println(err)
			return err
		}
	}
	return nil
}

func (p *Processor) finalizeAlreadySubmittedTransactionFromClientOrderlessChain(transaction *protos.Transaction) {
	var response *protos.TransactionResponse
	if transaction.Status == protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
		response = p.makeFailedTransactionResponse(transaction.TransactionId, transaction.Status, []byte("BlockHeader"))
	} else {
		response = p.makeSuccessTransactionResponse(transaction.TransactionId, []byte("BlockHeader"))
	}
	response.NodeSignature = p.signer.Sign(response.BlockHeader)
	p.sendTransactionResponseToSubscriber(transaction.ClientId, response)
}

func (p *Processor) finalizeTransactionResponseOrderlessChain(clientId string, txResponse *protos.TransactionResponse, transaction *protos.Transaction) {
	if !transaction.FromNode {
		txResponse.NodeSignature = p.signer.Sign(txResponse.BlockHeader)
		p.sendTransactionResponseToSubscriber(clientId, txResponse)
	}
	if txResponse.Status == protos.TransactionStatus_SUCCEEDED {
		p.transactionGossipListLock.Lock()
		p.transactionGossipList = append(p.transactionGossipList, transaction)
		p.transactionGossipListLock.Unlock()
	}
}

func (p *Processor) runGossipingOrderlessChain() {
	ticker := time.NewTicker(time.Duration(config.Config.GossipIntervalMS) * time.Millisecond)
	for range ticker.C {
		var transactionList []*protos.Transaction
		var transactionGossipListLength int
		p.transactionGossipListLock.Lock()
		transactionGossipListLength = len(p.transactionGossipList)
		if transactionGossipListLength > 0 {
			transactionList = make([]*protos.Transaction, transactionGossipListLength)
			copy(transactionList, p.transactionGossipList)
			p.transactionGossipList = []*protos.Transaction{}
		}
		p.transactionGossipListLock.Unlock()
		if transactionGossipListLength > 0 {
			go p.sendTransactionBatchToSubscribedNodeOrderlessChain(transactionList)
		}
	}
}

func (p *Processor) sendTransactionBatchToSubscribedNodeOrderlessChain(transactions []*protos.Transaction) {
	if p.ShouldFailByzantineNetwork() {
		return
	}
	gossips := &protos.NodeTransactionResponse{
		Transaction: transactions,
	}
	p.nodeSubscribersLock.RLock()
	nodes := reflect.ValueOf(p.OrderlessChainNodeTransactionResponseSubscriber).MapKeys()
	if profiling.IsBandwidthProfiling {
		_ = proto.Size(gossips)
	}
	p.nodeSubscribersLock.RUnlock()
	for _, node := range nodes {
		nodeId := node.String()
		p.nodeSubscribersLock.RLock()
		streamer, ok := p.OrderlessChainNodeTransactionResponseSubscriber[nodeId]
		p.nodeSubscribersLock.RUnlock()
		if !ok {
			logger.ErrorLogger.Println("Node was not found in the subscribers streams.", nodeId)
			return
		}
		if err := streamer.stream.Send(gossips); err != nil {
			streamer.finished <- true
			logger.ErrorLogger.Println("Could not send the response to the node " + nodeId)
			p.nodeSubscribersLock.Lock()
			delete(p.OrderlessChainNodeTransactionResponseSubscriber, nodeId)
			p.nodeSubscribersLock.Unlock()
		}
	}
}

func (p *Processor) NodeTransactionResponseSubscriptionOrderlessChain(subscription *protos.TransactionResponseEventSubscription,
	stream protos.TransactionService_SubscribeNodeTransactionsServer) error {
	finished := make(chan bool)
	p.nodeSubscribersLock.Lock()
	p.OrderlessChainNodeTransactionResponseSubscriber[subscription.ComponentId] = &OrderlessChainNodeTransactionResponseSubscriber{
		stream:   stream,
		finished: finished,
	}
	p.nodeSubscribersLock.Unlock()
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

func (p *Processor) subscriberForOtherNodeTransactionsOrderlessChain() {
	nodePseudoName := connpool.GetComponentPseudoName()
	for node := range p.gossipNodesConnectionPool {
		go func(node string) {
			for {
				conn, err := p.gossipNodesConnectionPool[node].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewTransactionServiceClient(conn.ClientConn)
				stream, err := client.SubscribeNodeTransactions(context.Background(), &protos.TransactionResponseEventSubscription{ComponentId: nodePseudoName})
				if err != nil {
					if errCon := conn.Close(); errCon != nil {
						logger.ErrorLogger.Println(errCon)
					}
					connpool.SleepAndReconnect()
					continue
				}
				for {
					if stream == nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					txPackage, streamErr := stream.Recv()
					if streamErr != nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					go p.processTransactionFromOtherNodesOrderlessChain(txPackage.Transaction)
				}
			}
		}(node)
	}
}
