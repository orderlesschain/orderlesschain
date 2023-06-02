package transactionprocessor

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transaction"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sort"
	"sync"
	"time"
)

func (p *Processor) signProposalResponseFabric(proposalResponse *protos.ProposalResponse) (*protos.ProposalResponse, error) {
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
	proposalResponse.NodeSignature = p.signer.Sign(marshalledReadWriteSet)
	return proposalResponse, nil
}

func (p *Processor) preProcessValidateReadWriteSetFabric(tx *protos.Transaction, wg *sync.WaitGroup) {
	sort.Slice(tx.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tx.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tx.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(tx.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return tx.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < tx.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	txDigest, err := proto.Marshal(tx.ReadWriteSet)
	if err != nil {
		tx.Status = protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION
		wg.Done()
		return
	}
	for node, nodeSign := range tx.NodeSignatures {
		err = p.signer.Verify(node, txDigest, nodeSign)
		if err != nil {
			tx.Status = protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION
			wg.Done()
			return
		}
	}
	wg.Done()
}

func (p *Processor) ProcessProposalFabricStream(proposal *protos.ProposalRequest) {
	p.transactionProfiler.AddEndorseStart(&proposal.ProposalId)
	p.txJournal.AddProposalToQueue(proposal)
}

func (p *Processor) runProposalQueueProcessingFabric() {
	for {
		proposals := <-p.txJournal.DequeuedProposalsChan
		go p.processDequeuedProposalsFabric(proposals)
	}
}

func (p *Processor) processDequeuedProposalsFabric(proposals *transaction.DequeuedProposals) {
	for _, proposal := range proposals.DequeuedProposals {
		go p.processProposalFabric(proposal)
	}
}

func (p *Processor) processProposalFabric(proposal *protos.ProposalRequest) {
	response, err := p.executeContract(proposal)
	if err != nil {
		p.sendProposalResponseToSubscriber(proposal.ClientId, response)
		return
	}
	response, err = p.signProposalResponseFabric(response)
	p.sendProposalResponseToSubscriber(proposal.ClientId, response)
}

func (p *Processor) subscriberForBlockEventsFabricAndFabricCRDT() {
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
					go p.transactionProfiler.AddCommitStartBlock(block)
					p.ProcessTransactionFabricAndFabricCRDTFromOrderer(block)
					go p.blockProcessor.AddBlockToDB(block)
				}
			}
		}(orderer)
	}
}

func (p *Processor) ProcessTransactionFabricAndFabricCRDTFromOrderer(block *protos.Block) {
	p.txJournal.AddBlockToQueue(block)
}

func (p *Processor) runTransactionProcessorFabricAndFabricCRDT() {
	for {
		p.txJournal.RequestBlockChan <- true
		block := <-p.txJournal.DequeuedOrdererBlockChan
		if block != nil {
			if config.Config.IsFabric {
				p.processBlockFabric(block)
			} else {
				p.processBlockFabricCRDT(block)
			}
		} else {
			time.Sleep(time.Duration(config.Config.QueueTickerDurationMS) * time.Millisecond)
		}
	}
}

func (p *Processor) processBlockFabric(block *protos.Block) {
	wg := &sync.WaitGroup{}
	wg.Add(len(block.Transactions))
	for _, tx := range block.Transactions {
		go p.preProcessValidateReadWriteSetFabric(tx, wg)
	}
	wg.Wait()
	for _, tx := range block.Transactions {
		p.processTransactionInBlockFabric(tx, block)
	}
}

func (p *Processor) processTransactionInBlockFabric(tx *protos.Transaction, block *protos.Block) {
	if tx.Status == protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
		go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION, block.ThisBlockHash))
	} else {
		err := p.validateMVCCFabric(tx)
		if err != nil {
			go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_MVCC, block.ThisBlockHash))
		}
		response, err := p.commitTransactionFabric(tx)
		if err == nil {
			response.BlockHeader = block.ThisBlockHash
			go p.sendTransactionResponseToSubscriber(tx.ClientId, response)
		} else {
			go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_DATABASE, block.ThisBlockHash))
		}
	}
}

func (p *Processor) validateMVCCFabric(tx *protos.Transaction) error {
	dbOp := p.sharedShimResources.DBConnections[tx.ContractName]
	for _, readKey := range tx.ReadWriteSet.ReadKeys.ReadKeys {
		if readKey.VersionNumber > 0 && dbOp.GetKeyCurrentVersion(readKey.Key) != readKey.VersionNumber {
			return errors.New("MVCC validation error")
		}
	}
	return nil
}

func (p *Processor) commitTransactionFabric(tx *protos.Transaction) (*protos.TransactionResponse, error) {
	if err := p.addTransactionToDatabaseWithVersion(tx); err != nil {
		return nil, err
	}
	return p.makeSuccessTransactionResponse(tx.TransactionId, []byte{}), nil
}
