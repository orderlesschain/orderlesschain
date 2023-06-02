package benchmark

import (
	"context"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/profiling"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sync"
	"time"
)

func (rex *RoundExecutor) executeTransactionPart1BIDL(counter int, startTime time.Time) {
	transactionResult := MakeNewTransactionResultFabricAndFabricCRDTAndBIDLAndSyncHotStuff(counter, rex.endorsementPolicyOrgs, rex.signer)
	rex.executor.transactionsResult.lock.Lock()
	rex.executor.transactionsResult.transactions[transactionResult.transaction.TransactionId] = transactionResult
	rex.executor.transactionsResult.lock.Unlock()
	proposal := transactionResult.transaction.MakeProposalRequestBenchmarkExecutor(transactionResult.transactionCounter, rex.baseContractOptions)
	proposal, _ = transactionResult.transaction.MakeProposalBIDLAndSyncHotStuffClientSign(proposal)
	transactionResult.readWriteType = proposal.WriteReadTransaction
	transactionResult.latencyMeasurementInstance = transactionResult.latencyMeasurement(startTime)
	rex.streamProposalBIDL(transactionResult, proposal)
}

func (rex *RoundExecutor) streamProposalBIDL(transactionResult *TransactionResult, transaction *protos.ProposalRequest) {
	for sequencer := range rex.executor.sequencersConnections {
		rex.executor.clientTransactionStreamLock.RLock()
		streamer, ok := rex.executor.clientTransactionStream[sequencer]
		rex.executor.clientTransactionStreamLock.RUnlock()
		if !ok {
			transactionResult.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			return
		}
		if err := streamer.streamBIDL.Send(transaction); err != nil {
			transactionResult.transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			rex.executor.clientTransactionStreamLock.Lock()
			delete(rex.executor.clientTransactionStream, sequencer)
			rex.executor.clientTransactionStreamLock.Unlock()
			err = rex.executor.makeSingleStreamTransactionSequencer(sequencer)
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
	}
}

func (ex *Executor) makeAllStreamTransactionBILD() {
	for sequencer := range ex.sequencersConnections {
		if err := ex.makeSingleStreamTransactionSequencer(sequencer); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}
}

func (ex *Executor) makeSingleStreamTransactionSequencer(sequencer string) error {
	conn, err := ex.sequencersConnections[sequencer].Get(context.Background())
	if conn == nil || err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	client := protos.NewSequencerServiceClient(conn.ClientConn)
	tempStream := &transactionStream{}
	tempStream.streamBIDL, err = client.BIDLTransactions(context.Background())
	if err != nil {
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(errCon)
		}
		connpool.SleepAndReconnect()
		err = ex.makeSingleStreamTransactionSequencer(sequencer)
		if err != nil {
			logger.ErrorLogger.Println(err)
		}
		return nil
	}
	ex.clientTransactionStreamLock.Lock()
	ex.clientTransactionStream[sequencer] = tempStream
	ex.clientTransactionStreamLock.Unlock()
	return nil
}

func (ex *Executor) subscriberForTransactionEventsBIDL() {
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
							ex.roundExecutor.executeTransactionPart2BIDL(tx)
						}
					}
				}
			}
		}(node)
	}
}

func (ex *Executor) addClientPublicKeys() {
	wg := &sync.WaitGroup{}
	wg.Add(len(ex.nodesConnectionsAddingPublicKeys))
	for name := range ex.nodesConnectionsAddingPublicKeys {
		go func(name string, wg *sync.WaitGroup) {
			conn, err := ex.nodesConnectionsAddingPublicKeys[name].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewTransactionServiceClient(conn.ClientConn)
			_, err = client.AddClientPublicKey(context.Background(), &protos.TransactionResponseEventSubscription{
				ComponentId: config.Config.UUID,
				PublicKey:   ex.PublicPrivateKey.PublicKeyString,
			})
			if err != nil {
				logger.ErrorLogger.Println(name, err)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(name, errCon)
			}
			wg.Done()
		}(name, wg)
	}
	wg.Wait()
	ex.nodesConnectionsAddingPublicKeys = nil
}

func (rex *RoundExecutor) executeTransactionPart2BIDL(tx *TransactionResult) {
	if tx.transaction.Status != protos.TransactionStatus_RUNNING {
		if tx.transaction.Status != protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
			logger.InfoLogger.Println(tx.transaction.Status)
		}
	} else {
		tx.transaction.Status = protos.TransactionStatus_SUCCEEDED
	}
	tx.EndTransactionMeasurements()
	rex.executor.makeTransactionDone()
}
