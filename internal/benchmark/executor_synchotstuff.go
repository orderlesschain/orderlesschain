package benchmark

import (
	"context"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"time"
)

func (rex *RoundExecutor) executeTransactionPart1SyncHotStuff(counter int, startTime time.Time) {
	transactionResult := MakeNewTransactionResultFabricAndFabricCRDTAndBIDLAndSyncHotStuff(counter, rex.endorsementPolicyOrgs, rex.signer)
	rex.executor.transactionsResult.lock.Lock()
	rex.executor.transactionsResult.transactions[transactionResult.transaction.TransactionId] = transactionResult
	rex.executor.transactionsResult.lock.Unlock()
	proposal := transactionResult.transaction.MakeProposalRequestBenchmarkExecutor(transactionResult.transactionCounter, rex.baseContractOptions)
	proposal, _ = transactionResult.transaction.MakeProposalBIDLAndSyncHotStuffClientSign(proposal)
	transactionResult.readWriteType = proposal.WriteReadTransaction
	transactionResult.latencyMeasurementInstance = transactionResult.latencyMeasurement(startTime)
	rex.streamProposalSyncHotStuff(proposal)
}

func (rex *RoundExecutor) streamProposalSyncHotStuff(transactionProposal *protos.ProposalRequest) {
	transaction := &protos.Transaction{
		TransactionId:   transactionProposal.ProposalId,
		ProposalRequest: transactionProposal,
	}
	for orderer := range rex.executor.orderersConnections {
		rex.executor.clientTransactionStreamLock.RLock()
		streamer, ok := rex.executor.clientTransactionStream[orderer]
		rex.executor.clientTransactionStreamLock.RUnlock()
		if !ok {
			transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			return
		}
		if err := streamer.streamSyncHotStuff.Send(transaction); err != nil {
			transaction.Status = protos.TransactionStatus_FAILED_GENERAL
			rex.executor.clientTransactionStreamLock.Lock()
			delete(rex.executor.clientTransactionStream, orderer)
			rex.executor.clientTransactionStreamLock.Unlock()
			if err = rex.executor.makeSingleTransactionStreamForLeader(orderer); err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
	}
}

func (ex *Executor) makeSyncHotStuffTransactionStreamForLeader() {
	for orderer := range ex.orderersConnections {
		if err := ex.makeSingleTransactionStreamForLeader(orderer); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}
}

func (ex *Executor) makeSingleTransactionStreamForLeader(orderer string) error {
	conn, err := ex.orderersConnections[orderer].Get(context.Background())
	if conn == nil || err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	client := protos.NewOrdererServiceClient(conn.ClientConn)
	tempStream := &transactionStream{}
	tempStream.streamSyncHotStuff, err = client.CommitSyncHotStuffTransactionStream(context.Background())
	if err != nil {
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(errCon)
		}
		connpool.SleepAndReconnect()
		err = ex.makeSingleTransactionStreamForLeader(orderer)
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
