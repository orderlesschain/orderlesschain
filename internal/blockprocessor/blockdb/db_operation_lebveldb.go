package blockdb

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
	"time"
)

const dbLevelDBLocation = "./data/blockchain_storage"

type BlockLevelDBOperations struct {
	db *leveldb.DB
}

func NewBlockLevelDBOperations() *BlockLevelDBOperations {
	tempDB, err := leveldb.OpenFile(dbLevelDBLocation, nil)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	return &BlockLevelDBOperations{
		db: tempDB,
	}
}

func (s *BlockLevelDBOperations) SaveNewBlock(block *protos.Block) error {
	marshalledBlock, err := proto.Marshal(block)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	if err = s.db.Put([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)), marshalledBlock, nil); err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	return nil
}

func (s *BlockLevelDBOperations) GetLastBlockHashAndSequence() ([]byte, int32, error) {
	iterator := s.db.NewIterator(nil, nil)
	var lastBlock []byte
	counter := 0
	for iterator.Next() {
		lastBlock = iterator.Value()
		counter++
	}
	iterator.Release()
	if counter == 0 {
		return nil, 0, errors.New("no block found")
	}
	if err := iterator.Error(); err != nil {
		return nil, 0, err
	}
	unMarshaledBlock := &protos.Block{}
	if err := proto.Unmarshal(lastBlock, unMarshaledBlock); err != nil {
		return nil, 0, err
	}
	return unMarshaledBlock.ThisBlockHash, unMarshaledBlock.BlockSequence, nil
}
