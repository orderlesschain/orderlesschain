package transactionprocessor

import (
	"context"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transaction"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"time"
)

func (p *Processor) verifyClientSignatureBIDLAndSyncHotStuff(proposalRequest *protos.ProposalRequest) error {
	marshalledProposal, _ := proto.Marshal(&protos.ProposalRequest{
		ProposalId:   proposalRequest.ProposalId,
		ClientId:     proposalRequest.ClientId,
		ContractName: proposalRequest.ContractName,
		MethodName:   proposalRequest.MethodName,
		MethodParams: proposalRequest.MethodParams,
	})
	if err := p.signer.Verify(proposalRequest.ClientId, marshalledProposal, proposalRequest.ClientSignature); err != nil {
		return err
	}
	return nil
}

func (p *Processor) runProposalQueueProcessingBIDL() {
	for {
		proposals := <-p.txJournal.DequeuedProposalsChan
		p.processDequeuedProposalsBIDL(proposals)
	}
}

func (p *Processor) processDequeuedProposalsBIDL(proposals *transaction.DequeuedProposals) {
	for _, proposal := range proposals.DequeuedProposals {
		p.processProposalBIDL(proposal)
	}
}

func (p *Processor) processProposalBIDL(proposal *protos.ProposalRequest) {
	if err := p.verifyClientSignatureBIDLAndSyncHotStuff(proposal); err != nil {
		logger.ErrorLogger.Println("client signature validation failed")
		go p.sendTransactionResponseToSubscriber(proposal.ClientId, p.makeFailedTransactionResponse(proposal.ProposalId,
			protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION, []byte{}))
		return
	}
	response, err := p.executeContract(proposal)
	if err != nil {
		go p.sendTransactionResponseToSubscriber(proposal.ClientId, p.makeFailedTransactionResponse(proposal.ProposalId,
			protos.TransactionStatus_FAILED_PROPOSAL, []byte{}))
		return
	}
	tx, err := p.txJournal.MakeTransactionFromProposalForBILDAnSyncHotStuff(proposal, response)
	if err != nil {
		go p.sendTransactionResponseToSubscriber(proposal.ClientId, p.makeFailedTransactionResponse(proposal.ProposalId,
			protos.TransactionStatus_FAILED_PROPOSAL, []byte{}))
		return
	}
	p.transactionProfiler.AddExecBidlEnd(&proposal.ProposalId)
	transactionToBeCommitted := p.txJournal.AddBILDTransactionFromSequencerAndCheckIfCanBeCommitted(tx)
	if transactionToBeCommitted != nil {
		p.transactionProfiler.AddCommitStart(&transactionToBeCommitted.TransactionId)
		p.processTransactionBIDLForCommit(transactionToBeCommitted)
	}
}

func (p *Processor) runTransactionProcessorBIDL() {
	for {
		p.txJournal.RequestBlockChan <- true
		block := <-p.txJournal.DequeuedOrdererBlockChan
		if block != nil {
			p.processBlockBIDL(block)
		} else {
			time.Sleep(time.Duration(config.Config.QueueTickerDurationMS) * time.Millisecond)
		}
	}
}

func (p *Processor) processBlockBIDL(block *protos.Block) {
	for _, tx := range block.Transactions {
		transactionToBeCommitted := p.txJournal.AddBILDTransactionFromOrdererAndCheckIfCanBeCommitted(tx)
		if transactionToBeCommitted != nil {
			p.transactionProfiler.AddCommitStart(&transactionToBeCommitted.TransactionId)
			p.processTransactionBIDLForCommit(transactionToBeCommitted)
		}
	}
}

func (p *Processor) processTransactionBIDLForCommit(tx *protos.Transaction) {
	response, err := p.commitTransactionBIDL(tx)
	if err == nil {
		response.BlockHeader = []byte{}
		go p.sendTransactionResponseToSubscriber(tx.ClientId, response)
	} else {
		go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_DATABASE, []byte{}))
	}
}

func (p *Processor) commitTransactionBIDL(tx *protos.Transaction) (*protos.TransactionResponse, error) {
	if err := p.addTransactionToDatabaseWithVersion(tx); err != nil {
		return nil, err
	}
	return p.makeSuccessTransactionResponse(tx.TransactionId, []byte{}), nil
}

func (p *Processor) addTransactionToDatabaseWithVersion(tx *protos.Transaction) error {
	dbOp := p.sharedShimResources.DBConnections[tx.ContractName]
	for _, keyValue := range tx.ReadWriteSet.WriteKeyValues.WriteKeyValues {
		if err := dbOp.PutKeyValueWithVersion(keyValue.Key, keyValue.Value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) subscriberForSequencedTransactions() {
	for _, sequencer := range p.inExperimentParticipatingSequencers {
		go func(sequencer string) {
			for {
				conn, err := p.sequencerConnectionPool[sequencer].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewSequencerServiceClient(conn.ClientConn)
				stream, err := client.SubscribeTransactionsForProcessing(context.Background(),
					&protos.SequencedTransactionForCommitting{NodeId: config.Config.UUID})
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
					transactionProposal, streamErr := stream.Recv()
					if streamErr != nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					p.transactionProfiler.AddExecBidlStart(&transactionProposal.ProposalId)
					p.txJournal.AddProposalToQueue(transactionProposal)
				}
			}
		}(sequencer)
	}
}

func (p *Processor) subscriberForBlockEventsBILD() {
	for _, orderer := range p.inExperimentParticipatingOrderers {
		go func(orderer string) {
			for {
				conn, err := p.ordererConnectionPool[orderer].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewOrdererServiceClient(conn.ClientConn)
				stream, err := client.SubscribeBlocks(context.Background(), &protos.BlockEventSubscription{NodeId: config.Config.UUID})
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
					block, streamErr := stream.Recv()
					if streamErr != nil {
						if errCon := conn.Close(); errCon != nil {
							logger.ErrorLogger.Println(errCon)
						}
						break
					}
					p.ProcessTransactionBIDLOrderer(block)
					go p.blockProcessor.AddBlockToDB(block)
				}
			}
		}(orderer)
	}
}

func (p *Processor) ProcessTransactionBIDLOrderer(block *protos.Block) {
	p.txJournal.AddBlockToQueue(block)
}
