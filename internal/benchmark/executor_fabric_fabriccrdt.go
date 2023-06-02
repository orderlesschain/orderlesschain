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

func (rex *RoundExecutor) executeTransactionPart1FabricAndFabricCRDT(counter int, startTime time.Time) {
	transactionResult := MakeNewTransactionResultFabricAndFabricCRDTAndBIDLAndSyncHotStuff(counter, rex.endorsementPolicyOrgs, rex.signer)
	rex.executor.transactionsResult.lock.Lock()
	rex.executor.transactionsResult.transactions[transactionResult.transaction.TransactionId] = transactionResult
	rex.executor.transactionsResult.lock.Unlock()
	proposal := transactionResult.transaction.MakeProposalRequestBenchmarkExecutor(transactionResult.transactionCounter, rex.baseContractOptions)
	transactionResult.readWriteType = proposal.WriteReadTransaction
	transactionResult.latencyMeasurementInstance = transactionResult.latencyMeasurement(startTime)
	rex.streamProposalFabricAndFabricCRDT(transactionResult, proposal)
}

func (rex *RoundExecutor) executeTransactionPart2FabricAndFabricCRDT(tx *TransactionResult) {
	if tx.transaction.Status == protos.TransactionStatus_FAILED_GENERAL {
		tx.EndTransactionMeasurements()
		rex.executor.makeTransactionDone()
		return
	}
	var commitTransaction *protos.Transaction
	var err error
	if config.Config.IsFabric {
		commitTransaction, err = tx.transaction.MakeTransactionBenchmarkExecutorWithClientFullSign(rex.benchmarkConfig, rex.endorsementPolicyOrgs)
	} else {
		commitTransaction, err = tx.transaction.MakeTransactionBenchmarkExecutorWithClientBaseSign(rex.benchmarkConfig, rex.endorsementPolicyOrgs)
	}
	if err != nil {
		tx.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
		tx.EndTransactionMeasurements()
		rex.executor.makeTransactionDone()
		return
	}
	rex.streamTransactionFabricAndFabricCRDT(tx, commitTransaction)
	if tx.transaction.Status == protos.TransactionStatus_FAILED_GENERAL {
		tx.EndTransactionMeasurements()
		rex.executor.makeTransactionDone()
		return
	}
}

func (rex *RoundExecutor) executeTransactionPart3FabricAndFabricCRDT(tx *TransactionResult) {
	if tx.transaction.Status != protos.TransactionStatus_RUNNING {
		if tx.transaction.Status != protos.TransactionStatus_FAILED_MVCC &&
			tx.transaction.Status != protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
		}
	} else {
		tx.transaction.Status = protos.TransactionStatus_SUCCEEDED
	}
	tx.EndTransactionMeasurements()
	rex.executor.makeTransactionDone()
}

func (rex *RoundExecutor) streamProposalFabricAndFabricCRDT(transaction *TransactionResult, proposal *protos.ProposalRequest) {
	if profiling.IsBandwidthProfiling {
		transaction.sentProposalBytes = rex.selectedOrgsEndorsementPolicyCount * proto.Size(proposal)
	}
	for _, nodeId := range rex.selectedOrgsEndorsementPolicy {
		rex.executor.clientProposalStreamLock.RLock()
		streamer, ok := rex.executor.clientProposalStream[nodeId]
		rex.executor.clientProposalStreamLock.RUnlock()
		if !ok {
			transaction.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			return
		}
		var err error
		if config.Config.IsFabric {
			err = streamer.streamFabric.Send(proposal)
		} else {
			err = streamer.streamFabricCRDT.Send(proposal)
		}
		if err != nil {
			transaction.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			rex.executor.clientProposalStreamLock.Lock()
			delete(rex.executor.clientProposalStream, nodeId)
			rex.executor.clientProposalStreamLock.Unlock()
			err = rex.executor.makeSingleStreamProposal(nodeId)
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
	}
}

func (rex *RoundExecutor) streamTransactionFabricAndFabricCRDT(transactionResult *TransactionResult, transaction *protos.Transaction) {
	if profiling.IsBandwidthProfiling {
		transactionResult.sentTransactionBytes = proto.Size(transaction)
	}
	for orderer := range rex.executor.orderersConnections {
		rex.executor.clientTransactionStreamLock.RLock()
		streamer, ok := rex.executor.clientTransactionStream[orderer]
		rex.executor.clientTransactionStreamLock.RUnlock()
		if !ok {
			transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			return
		}
		if err := streamer.streamFabricAndFabricCRDT.Send(transaction); err != nil {
			transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			rex.executor.clientTransactionStreamLock.Lock()
			delete(rex.executor.clientTransactionStream, orderer)
			rex.executor.clientTransactionStreamLock.Unlock()
			err = rex.executor.makeSingleStreamTransactionOrderer(orderer)
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
	}
}

func (ex *Executor) subscriberForProposalEventsFabricAndFabricCRDT() {
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
					if ex.roundNotDone {
						readyToSendTransaction := false
						ex.transactionsResult.lock.Lock()
						tx := ex.transactionsResult.transactions[proposalResponse.ProposalId]
						tx.receivedProposalCount++
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
							tx.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
						}
						if readyToSendTransaction {
							go ex.roundExecutor.executeTransactionPart2FabricAndFabricCRDT(tx)
						}
					}
				}
			}
		}(node)
	}
}

func (ex *Executor) subscriberForTransactionEventsFabricAndFabricCRDTAndSyncHotStuff() {
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
					if ex.roundNotDone {
						readyToConcludeTransaction := false
						ex.transactionsResult.lock.Lock()
						tx := ex.transactionsResult.transactions[txResponse.TransactionId]
						tx.receivedTransactionCommitCount++
						if tx.receivedTransactionCommitCount <= tx.receivedTransactionCommitExpected {
							if tx.receivedTransactionCommitCount == tx.receivedTransactionCommitExpected {
								readyToConcludeTransaction = true
							}
						}
						ex.transactionsResult.lock.Unlock()
						if profiling.IsBandwidthProfiling {
							tx.receivedTransactionBytes += proto.Size(txResponse)
						}
						if txResponse.Status != protos.TransactionStatus_SUCCEEDED {
							tx.transaction.Status = txResponse.Status
						}
						if readyToConcludeTransaction {
							ex.roundExecutor.executeTransactionPart3FabricAndFabricCRDT(tx)
						}
					}
				}
			}
		}(node)
	}
}

func (ex *Executor) makeAllStreamTransactionFabricAndFabricCRDT() {
	for orderer := range ex.orderersConnections {
		if err := ex.makeSingleStreamTransactionOrderer(orderer); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}
}

func (ex *Executor) makeSingleStreamTransactionOrderer(orderer string) error {
	conn, err := ex.orderersConnections[orderer].Get(context.Background())
	if conn == nil || err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	client := protos.NewOrdererServiceClient(conn.ClientConn)
	tempStream := &transactionStream{}
	tempStream.streamFabricAndFabricCRDT, err = client.CommitFabricAndFabricCRDTTransactionStream(context.Background())
	if err != nil {
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(errCon)
		}
		connpool.SleepAndReconnect()
		err = ex.makeSingleStreamTransactionOrderer(orderer)
		if err != nil {
			logger.ErrorLogger.Println(err)
		}
		return nil
	}
	ex.clientTransactionStreamLock.Lock()
	ex.clientTransactionStream[orderer] = tempStream
	ex.clientTransactionStreamLock.Unlock()
	return nil
}
