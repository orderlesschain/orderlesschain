package blockprocessor

import (
	"github.com/google/uuid"
	"gitlab.lrz.de/orderless/orderlesschain/internal/blockprocessor/blockdb"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/hasher"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/keygenerator"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/signer"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"time"
)

type MinedBlock struct {
	Transactions []*protos.Transaction
	Block        *protos.Block
}

type Processor struct {
	dbConnections         *blockdb.BlockLevelDBOperations
	TransactionChan       chan *protos.Transaction
	currentBlock          *protos.Block
	currentBlockMetadata  *MinedBlock
	MinedBlock            chan *MinedBlock
	minedBlockStorageChan chan *protos.Block
	signer                *signer.Signer
}

func InitBlockProcessor() *Processor {
	publicPrivateKey := keygenerator.LoadPublicPrivateKeyFromFile()
	tempProcessor := &Processor{
		dbConnections:         blockdb.NewBlockLevelDBOperations(),
		TransactionChan:       make(chan *protos.Transaction),
		MinedBlock:            make(chan *MinedBlock),
		minedBlockStorageChan: make(chan *protos.Block),
		signer:                signer.NewSigner(publicPrivateKey),
	}
	previousBlockHash, blockSequence, err := tempProcessor.dbConnections.GetLastBlockHashAndSequence()
	if err != nil {
		previousBlockHash = []byte("GENESIS BLOCK")
		blockSequence = 1
	}
	tempProcessor.currentBlock = newBlock(previousBlockHash, blockSequence)
	tempProcessor.currentBlockMetadata = newBlockMetadata()
	go tempProcessor.runBlockProcessor()
	go tempProcessor.runBlockStorage()
	return tempProcessor
}

func newBlock(previousBlockHash []byte, blockSequence int32) *protos.Block {
	return &protos.Block{
		BlockId:            uuid.NewString(),
		ThisBlockHash:      []byte{},
		PreviousBlockHash:  previousBlockHash,
		TransactionsDigest: []byte{},
		Transactions:       make([]*protos.Transaction, 0, config.Config.BlockTransactionSize),
		BlockSequence:      blockSequence,
	}
}

func newBlockMetadata() *MinedBlock {
	return &MinedBlock{
		Transactions: make([]*protos.Transaction, 0, config.Config.BlockTransactionSize),
	}
}

func (p *Processor) runBlockProcessor() {
	var timer <-chan time.Time
	duration := time.Duration(config.Config.BlockTimeoutMS) * time.Millisecond
	var nodeSign []byte
	for {
		select {
		case tx := <-p.TransactionChan:
			p.currentBlock.Transactions = append(p.currentBlock.Transactions, tx)
			p.currentBlockMetadata.Transactions = append(p.currentBlockMetadata.Transactions, tx)
			for _, nodeSign = range tx.NodeSignatures {
				break
			}
			if len(nodeSign) >= hasher.Limit {
				p.currentBlock.TransactionsDigest = append(p.currentBlock.TransactionsDigest, nodeSign[:hasher.Limit]...)
			}
			if len(p.currentBlock.Transactions) == config.Config.BlockTransactionSize {
				p.commitBlock()
				timer = nil
			} else {
				timer = time.After(duration)
			}
		case <-timer:
			timer = nil
			if len(p.currentBlock.Transactions) == 0 {
				continue
			} else {
				p.commitBlock()
			}
		}
	}
}

func (p *Processor) commitBlock() {
	createdBlock, err := p.prepareBlockCommit()
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	p.currentBlockMetadata.Block = createdBlock
	p.MinedBlock <- p.currentBlockMetadata
	p.currentBlock = newBlock(createdBlock.ThisBlockHash, createdBlock.BlockSequence+1)
	p.currentBlockMetadata = newBlockMetadata()
	go p.AddBlockToDB(createdBlock)
}

func (p *Processor) prepareBlockCommit() (*protos.Block, error) {
	p.currentBlock.TransactionsDigest = append(p.currentBlock.TransactionsDigest, p.currentBlock.PreviousBlockHash...)
	if len(p.currentBlock.TransactionsDigest) >= hasher.LimitHashedBlock {
		p.currentBlock.ThisBlockHash = hasher.Hash(p.currentBlock.TransactionsDigest[:hasher.LimitHashedBlock])
	} else {
		p.currentBlock.ThisBlockHash = hasher.Hash(p.currentBlock.TransactionsDigest)
	}
	return p.currentBlock, nil
}

func (p *Processor) AddBlockToDB(block *protos.Block) {
	p.minedBlockStorageChan <- block
}

func (p *Processor) runBlockStorage() {
	for {
		err := p.dbConnections.SaveNewBlock(<-p.minedBlockStorageChan)
		if err != nil {
			logger.FatalLogger.Fatalln(err)
		}
	}
}
