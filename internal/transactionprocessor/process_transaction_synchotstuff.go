package transactionprocessor

import (
	"context"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"reflect"
)

func (p *Processor) subscriberForBlockEventsSyncHotStuff() {
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
					go p.transactionProfiler.AddEndorseStartBlock(block)
					p.txJournal.AddBlockToQueueSyncHotStuff(block)
					go p.sendSignedBlockToOtherNodesSyncHotStuff(block)
				}
			}
		}(orderer)
	}
}

func (p *Processor) BlockSubscriptionFromOtherNodesSyncHotStuff(subscriberEvent *protos.BlockEventSubscription,
	stream protos.TransactionService_SubscribeBlocksSyncHotStuffServer) error {
	finished := make(chan bool)
	p.subscribersLockSyncHotStuff.Lock()
	p.blockSubscribersSyncHotStuff[subscriberEvent.NodeId] = &blockResponseSubscriberSyncHotStuff{
		stream:   stream,
		finished: finished,
	}
	p.subscribersLockSyncHotStuff.Unlock()
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

func (p *Processor) signBlockSyncHotStuff(block *protos.Block) (*protos.Block, error) {
	signedBlock := &protos.Block{
		BlockId:                block.BlockId,
		ThisBlockHash:          block.ThisBlockHash,
		SignerNodeSynchotstuff: config.Config.UUID,
		QuorumSyncHotStuff:     block.Transactions[0].EndorsementPolicy,
	}
	signedBlock.BlockSignatureSynchotstuff = p.signer.Sign(block.ThisBlockHash)
	return signedBlock, nil
}

func (p *Processor) sendSignedBlockToOtherNodesSyncHotStuff(unsignedBlock *protos.Block) {
	signedBlock, _ := p.signBlockSyncHotStuff(unsignedBlock)
	p.subscribersLockSyncHotStuff.RLock()
	nodes := reflect.ValueOf(p.blockSubscribersSyncHotStuff).MapKeys()
	p.subscribersLockSyncHotStuff.RUnlock()
	for _, node := range nodes {
		nodeId := node.String()
		p.subscribersLockSyncHotStuff.RLock()
		streamer, ok := p.blockSubscribersSyncHotStuff[nodeId]
		p.subscribersLockSyncHotStuff.RUnlock()
		if !ok {
			logger.ErrorLogger.Println("Node was not found in the subscribers streams.", nodeId)
			return
		}
		if err := streamer.stream.Send(signedBlock); err != nil {
			streamer.finished <- true
			logger.ErrorLogger.Println("Could not send the response to the node " + nodeId)
			p.subscribersLockSyncHotStuff.Lock()
			delete(p.blockSubscribersSyncHotStuff, nodeId)
			p.subscribersLockSyncHotStuff.Unlock()
		}
	}
}

func (p *Processor) subscribeForOtherNodeBlockSyncHotStuff() {
	for node := range p.syncHotStuffNodesConnectionPool {
		go func(node string) {
			for {
				conn, err := p.syncHotStuffNodesConnectionPool[node].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewTransactionServiceClient(conn.ClientConn)
				stream, err := client.SubscribeBlocksSyncHotStuff(context.Background(), &protos.BlockEventSubscription{NodeId: config.Config.UUID})
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
					go p.verifyBlockReceivedFromNode(block)
				}
			}
		}(node)
	}
}

func (p *Processor) verifyBlockReceivedFromNode(block *protos.Block) {
	err := p.signer.Verify(block.SignerNodeSynchotstuff, block.ThisBlockHash, block.BlockSignatureSynchotstuff)
	if err != nil {
		logger.ErrorLogger.Println("block signature validation failed")
		return
	}
	readyBlock := p.txJournal.AddNodeVoteSyncHotStuff(block.BlockId, block.SignerNodeSynchotstuff, int(block.QuorumSyncHotStuff))
	if readyBlock != nil {
		go p.transactionProfiler.AddEndorseEndBlock(readyBlock)
		go p.transactionProfiler.AddCommitStartBlock(readyBlock)
		p.readyBlockSyncHotStuff <- readyBlock
		go p.blockProcessor.AddBlockToDB(readyBlock)
	}
}

func (p *Processor) runProcessBlockSyncHotStuff() {
	for {
		block := <-p.readyBlockSyncHotStuff
		p.processBlockSyncHotStuff(block)
	}
}

func (p *Processor) processBlockSyncHotStuff(block *protos.Block) {
	for _, tx := range block.Transactions {
		p.processTransactionSyncHotStuff(tx)
	}
}

func (p *Processor) processTransactionSyncHotStuff(transaction *protos.Transaction) {
	if err := p.verifyClientSignatureBIDLAndSyncHotStuff(transaction.ProposalRequest); err != nil {
		logger.ErrorLogger.Println("client signature validation failed")
		go p.sendTransactionResponseToSubscriber(transaction.ClientId, p.makeFailedTransactionResponse(transaction.TransactionId,
			protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION, []byte{}))
		return
	}
	response, err := p.executeContract(transaction.ProposalRequest)
	if err != nil {
		go p.sendTransactionResponseToSubscriber(transaction.ClientId, p.makeFailedTransactionResponse(transaction.TransactionId,
			protos.TransactionStatus_FAILED_PROPOSAL, []byte{}))
		return
	}
	tx, err := p.txJournal.MakeTransactionFromProposalForBILDAnSyncHotStuff(transaction.ProposalRequest, response)
	if err != nil {
		go p.sendTransactionResponseToSubscriber(transaction.ClientId, p.makeFailedTransactionResponse(transaction.TransactionId,
			protos.TransactionStatus_FAILED_PROPOSAL, []byte{}))
		return
	}
	p.processTransactionSyncHotStuffForCommit(tx)
}

func (p *Processor) processTransactionSyncHotStuffForCommit(tx *protos.Transaction) {
	response, err := p.commitTransactionSyncHotStuff(tx)
	if err == nil {
		response.BlockHeader = []byte{}
		go p.sendTransactionResponseToSubscriber(tx.ClientId, response)
	} else {
		go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_DATABASE, []byte{}))
	}
}

func (p *Processor) commitTransactionSyncHotStuff(tx *protos.Transaction) (*protos.TransactionResponse, error) {
	if err := p.addTransactionToDatabaseWithVersion(tx); err != nil {
		return nil, err
	}
	return p.makeSuccessTransactionResponse(tx.TransactionId, []byte{}), nil
}
