package benchmark

import (
	"context"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/profiling"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"time"
)

func (rex *RoundExecutor) executeTransactionPart1OrderlessChain(counter int, startTime time.Time) {
	transactionResult := MakeNewTransactionResultOrderlessChain(counter, rex.endorsementPolicyOrgsWithExtraEndorsement, rex.endorsementPolicyOrgs, rex.signer)
	rex.executor.transactionsResult.lock.Lock()
	rex.executor.transactionsResult.transactions[transactionResult.transaction.TransactionId] = transactionResult
	rex.executor.transactionsResult.lock.Unlock()
	proposal := transactionResult.transaction.MakeProposalRequestBenchmarkExecutor(transactionResult.transactionCounter, rex.baseContractOptions)
	transactionResult.readWriteType = proposal.WriteReadTransaction
	transactionResult.latencyMeasurementInstance = transactionResult.latencyMeasurement(startTime)
	rex.streamProposalOrderlessChain(transactionResult, proposal)
}

func (rex *RoundExecutor) executeTransactionPart2OrderlessChain(tx *TransactionResult) {
	if rex.executor.ShouldFailByzantineNetwork() {
		tx.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
	}
	if tx.transaction.Status == protos.TransactionStatus_FAILED_GENERAL {
		tx.EndTransactionMeasurements()
		rex.executor.makeTransactionDone()
		return
	}
	var commitTransaction *protos.Transaction
	var err error
	if rex.executor.ShouldFailByzantineTampered() {
		commitTransaction, err = tx.transaction.MakeByzantineTransactionBenchmarkExecutorWithClientFullSign(rex.benchmarkConfig, rex.endorsementPolicyOrgs)
	} else {
		commitTransaction, err = tx.transaction.MakeTransactionBenchmarkExecutorWithClientFullSign(rex.benchmarkConfig, rex.endorsementPolicyOrgs)
	}
	if err != nil {
		tx.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
		tx.EndTransactionMeasurements()
		rex.executor.makeTransactionDone()
		return
	}
	rex.streamTransactionOrderlessChain(tx, commitTransaction)
	if tx.transaction.Status == protos.TransactionStatus_FAILED_GENERAL {
		tx.EndTransactionMeasurements()
		rex.executor.makeTransactionDone()
		return
	}
}

func (rex *RoundExecutor) executeTransactionPart3OrderlessChain(tx *TransactionResult) {
	if tx.transaction.Status == protos.TransactionStatus_RUNNING {
		tx.transaction.Status = protos.TransactionStatus_SUCCEEDED
	}
	tx.EndTransactionMeasurements()
	rex.executor.makeTransactionDone()
}

func (rex *RoundExecutor) streamProposalOrderlessChain(transaction *TransactionResult, proposal *protos.ProposalRequest) {
	if profiling.IsBandwidthProfiling {
		transaction.sentProposalBytes = rex.selectedOrgsEndorsementPolicyCount * proto.Size(proposal)
	}
	sentProposals := 0
	for _, nodeId := range rex.selectedOrgsEndorsementPolicy {
		rex.executor.clientProposalStreamLock.RLock()
		streamer, ok := rex.executor.clientProposalStream[nodeId]
		rex.executor.clientProposalStreamLock.RUnlock()
		if !ok {
			continue
		}
		if err := streamer.streamOrderlessChain.Send(proposal); err != nil {
			rex.executor.clientProposalStreamLock.Lock()
			delete(rex.executor.clientProposalStream, nodeId)
			rex.executor.clientProposalStreamLock.Unlock()
			err = rex.executor.makeSingleStreamProposal(nodeId)
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
		sentProposals++

	}
	if sentProposals < rex.endorsementPolicyOrgs {
		transaction.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
	}
}

func (rex *RoundExecutor) streamTransactionOrderlessChain(transactionResult *TransactionResult, transaction *protos.Transaction) {
	if profiling.IsBandwidthProfiling {
		transactionResult.sentTransactionBytes = rex.selectedOrgsEndorsementPolicyCount * proto.Size(transaction)
	}
	sentTransactions := 0
	for _, nodeId := range rex.selectedOrgsEndorsementPolicy {
		rex.executor.clientTransactionStreamLock.RLock()
		streamer, ok := rex.executor.clientTransactionStream[nodeId]
		rex.executor.clientTransactionStreamLock.RUnlock()
		if !ok {
			continue
		}
		if err := streamer.streamOrderlessChain.Send(transaction); err != nil {
			rex.executor.clientTransactionStreamLock.Lock()
			delete(rex.executor.clientTransactionStream, nodeId)
			rex.executor.clientTransactionStreamLock.Unlock()
			err = rex.executor.makeSingleStreamTransactionOrderlessChain(nodeId)
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
		sentTransactions++
	}
	if sentTransactions < rex.endorsementPolicyOrgs {
		transaction.Status = protos.TransactionStatus_FAILED_GENERAL
	}
}

func (ex *Executor) subscriberForProposalEventsOrderlessChain() {
	for node := range ex.nodesConnectionsWatchProposalEvent {
		go func(node string) {
			for {
				conn, err := ex.nodesConnectionsWatchProposalEvent[node].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewTransactionServiceClient(conn.ClientConn)
				stream, err := client.SubscribeProposalResponse(context.Background(), &protos.ProposalResponseEventSubscription{ComponentId: config.Config.UUID})
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
					proposalResponse, streamErr := stream.Recv()
					if streamErr != nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					ex.processProposalResponse(proposalResponse)
				}
			}
		}(node)
	}
}

func (ex *Executor) subscriberForNewlyAddedProposalEventsOrderlessChain(newNodes map[string]bool) {
	for node := range newNodes {
		go func(node string) {
			for {
				conn, err := ex.nodesConnectionsWatchProposalEvent[node].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewTransactionServiceClient(conn.ClientConn)
				stream, err := client.SubscribeProposalResponse(context.Background(), &protos.ProposalResponseEventSubscription{ComponentId: config.Config.UUID})
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
					proposalResponse, streamErr := stream.Recv()
					if streamErr != nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					ex.processProposalResponse(proposalResponse)
				}
			}
		}(node)
	}
}

func (ex *Executor) processProposalResponse(proposalResponse *protos.ProposalResponse) {
	if ex.roundNotDone {
		readyToSendTransaction := false
		ex.transactionsResult.lock.Lock()
		tx := ex.transactionsResult.transactions[proposalResponse.ProposalId]
		if proposalResponse.Status == protos.ProposalResponse_SUCCESS {
			tx.receivedProposalCount++
		}
		if tx.receivedProposalCount <= tx.receivedProposalExpected {
			tx.transaction.ProposalResponses[proposalResponse.NodeId] = proposalResponse
			if tx.receivedProposalCount == tx.receivedProposalExpected {
				readyToSendTransaction = true
			}
		}
		ex.transactionsResult.lock.Unlock()
		if profiling.IsBandwidthProfiling {
			tx.receivedProposalBytes += proto.Size(proposalResponse)
		}
		if proposalResponse.Status != protos.ProposalResponse_SUCCESS {
			logger.InfoLogger.Println("Transaction failed", protos.TransactionStatus_FAILED_GENERAL)
		}
		if readyToSendTransaction {
			go ex.roundExecutor.executeTransactionPart2OrderlessChain(tx)
		}
	}
}

func (ex *Executor) subscriberForTransactionEventsOrderlessChain() {
	for node := range ex.nodesConnectionsWatchTransactionEvent {
		go func(node string) {
			for {
				conn, err := ex.nodesConnectionsWatchTransactionEvent[node].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewTransactionServiceClient(conn.ClientConn)
				stream, err := client.SubscribeTransactionResponse(context.Background(), &protos.TransactionResponseEventSubscription{
					ComponentId: config.Config.UUID,
					PublicKey:   ex.PublicPrivateKey.PublicKeyString,
				})
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
					txResponse, streamErr := stream.Recv()
					if streamErr != nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					ex.processTransactionResponse(txResponse)
				}
			}
		}(node)
	}
}

func (ex *Executor) subscriberForNewlyAddedTransactionEventsOrderlessChain(newNodes map[string]bool) {
	for node := range newNodes {
		go func(node string) {
			for {
				conn, err := ex.nodesConnectionsWatchTransactionEvent[node].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewTransactionServiceClient(conn.ClientConn)
				stream, err := client.SubscribeTransactionResponse(context.Background(), &protos.TransactionResponseEventSubscription{
					ComponentId: config.Config.UUID,
					PublicKey:   ex.PublicPrivateKey.PublicKeyString,
				})
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
					txResponse, streamErr := stream.Recv()
					if streamErr != nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					ex.processTransactionResponse(txResponse)
				}
			}
		}(node)
	}
}

func (ex *Executor) processTransactionResponse(txResponse *protos.TransactionResponse) {
	if ex.roundNotDone {
		readyToConcludeTransaction := false
		ex.transactionsResult.lock.Lock()
		tx := ex.transactionsResult.transactions[txResponse.TransactionId]
		if txResponse.Status == protos.TransactionStatus_SUCCEEDED {
			tx.receivedTransactionCommitCount++
		}
		if tx.receivedTransactionCommitCount <= tx.receivedTransactionCommitExpected {
			if tx.receivedTransactionCommitCount == tx.receivedTransactionCommitExpected {
				readyToConcludeTransaction = true
			}
		}
		ex.transactionsResult.lock.Unlock()
		if profiling.IsBandwidthProfiling {
			tx.receivedTransactionBytes += proto.Size(txResponse)
		}
		if readyToConcludeTransaction {
			ex.roundExecutor.executeTransactionPart3OrderlessChain(tx)
		}
	}
}

func (ex *Executor) makeAllStreamTransactionOrderlessChain() {
	for node := range ex.nodesConnectionsStreamTransactions {
		if err := ex.makeSingleStreamTransactionOrderlessChain(node); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}
}

func (ex *Executor) makeSingleStreamTransactionOrderlessChain(node string) error {
	conn, err := ex.nodesConnectionsStreamTransactions[node].Get(context.Background())
	if conn == nil || err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	client := protos.NewTransactionServiceClient(conn.ClientConn)
	tempStream := &transactionStream{}
	tempStream.streamOrderlessChain, err = client.CommitOrderlessChainTransactionStream(context.Background())
	if err != nil {
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(errCon)
		}
		connpool.SleepAndReconnect()
		err = ex.makeSingleStreamTransactionOrderlessChain(node)
		if err != nil {
			logger.ErrorLogger.Println(err)
		}
		return nil
	}
	ex.clientTransactionStreamLock.Lock()
	ex.clientTransactionStream[node] = tempStream
	ex.clientTransactionStreamLock.Unlock()
	return nil
}
