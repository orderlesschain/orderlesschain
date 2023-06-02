package transactiondb

import (
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
)

const dbsLocation = "./data/state_"

type Operations struct {
	db *leveldb.DB
}

type ByteContainer struct {
	Value []byte
}

type ByteContainers struct {
	Values []*ByteContainer
}

// NewOperations Possible optimization for LevelDB: https://github.com/google/leveldb/blob/master/doc/index.md
// Comparison of LevelDB to SQLite http://www.lmdb.tech/bench/microbench/benchmark.html
func NewOperations(contractName string) *Operations {
	tempDB, err := leveldb.OpenFile(dbsLocation+contractName, nil)
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
	return &Operations{
		db: tempDB,
	}
}

func (o *Operations) GetKeyValueWithVersion(key string) (*protos.WriteValue, error) {
	if valueBytes, err := o.db.Get([]byte(key), nil); err != nil {
		return nil, err
	} else {
		value := &protos.WriteValue{}
		err = proto.Unmarshal(valueBytes, value)
		return value, err
	}
}

func (o *Operations) GetKeyRangeValueWithVersion(key string) (map[string]*protos.WriteValue, error) {
	keyValues := map[string]*protos.WriteValue{}
	iter := o.db.NewIterator(util.BytesPrefix([]byte(key+"-")), nil)
	for iter.Next() {
		value := &protos.WriteValue{}
		err := proto.Unmarshal(iter.Value(), value)
		if err != nil {
			return nil, err
		}
		keyValues[string(iter.Key())] = value
	}
	iter.Release()
	return keyValues, iter.Error()
}

func (o *Operations) GetKeyCurrentVersion(key string) int32 {
	if valueBytes, err := o.db.Get([]byte(key), nil); err != nil {
		return 0
	} else {
		value := &protos.WriteValue{}
		err = proto.Unmarshal(valueBytes, value)
		return value.VersionNumber
	}
}

func (o *Operations) PutKeyValueWithVersion(key string, value []byte) error {
	updateValue := &protos.WriteValue{}
	if valueBytes, err := o.db.Get([]byte(key), nil); err == nil {
		oldValue := &protos.WriteValue{}
		err = proto.Unmarshal(valueBytes, oldValue)
		updateValue.VersionNumber = oldValue.VersionNumber
	}
	updateValue.VersionNumber++
	updateValue.Value = value
	dbValue, err := proto.Marshal(updateValue)
	err = o.db.Put([]byte(key), dbValue, nil)
	return err
}

func (o *Operations) GetKeyValueNoVersion(key string) ([]byte, error) {
	if valueBytes, err := o.db.Get([]byte(key), nil); err != nil {
		return []byte{}, err
	} else {
		return valueBytes, nil
	}
}

func (o *Operations) GetKeyRangeNoVersion(key string) (map[string][]byte, error) {
	keyValues := map[string][]byte{}
	iter := o.db.NewIterator(util.BytesPrefix([]byte(key+"-")), nil)
	for iter.Next() {
		keyValues[string(iter.Key())] = iter.Value()
	}
	iter.Release()
	return keyValues, iter.Error()
}

func (o *Operations) PutKeyValueNoVersion(key string, value []byte) error {
	err := o.db.Put([]byte(key), value, nil)
	return err
}

func (o *Operations) GetCRDTOperationsV1(key string) (*protos.DocChanges, error) {
	docChanges := &protos.DocChanges{}
	docChanges.Operations = []*protos.Operation{}
	iter := o.db.NewIterator(util.BytesPrefix([]byte(key+"-")), nil)
	for iter.Next() {
		docChange := &protos.DocChanges{}
		err := proto.Unmarshal(iter.Value(), docChange)
		if err != nil {
			return &protos.DocChanges{}, err
		}
		docChanges.Operations = append(docChanges.Operations, docChange.Operations...)
	}
	iter.Release()
	return docChanges, iter.Error()
}

func (o *Operations) GetCRDTOperationsV2(key string) (*protos.CRDTOperationsList, error) {
	crdtOperationsList := &protos.CRDTOperationsList{}
	crdtOperationsList.CrdtObjectId = key
	crdtOperationsList.Operations = []*protos.CRDTOperation{}
	iter := o.db.NewIterator(util.BytesPrefix([]byte(key+"-")), nil)
	for iter.Next() {
		tempCrdtOperationsList := &protos.CRDTOperationsList{}
		err := proto.Unmarshal(iter.Value(), tempCrdtOperationsList)
		if err != nil {
			return &protos.CRDTOperationsList{}, err
		}
		crdtOperationsList.Operations = append(crdtOperationsList.Operations, tempCrdtOperationsList.Operations...)
	}
	iter.Release()
	return crdtOperationsList, iter.Error()
}
