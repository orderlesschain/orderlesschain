package crdtmanagerv2

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor/transactiondb"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strings"
	"sync"
)

type Manager struct {
	crdtObjs      map[string]*CRDTObject
	DBConnections map[string]*transactiondb.Operations
	lock          *sync.RWMutex
}

func NewManager(dbConnections map[string]*transactiondb.Operations) *Manager {
	return &Manager{
		crdtObjs:      map[string]*CRDTObject{},
		DBConnections: dbConnections,
		lock:          &sync.RWMutex{},
	}
}

type CRDTObject struct {
	CrdtObj *protos.CRDTObject
	Lock    *sync.RWMutex
}

type CRDTOperationsPut struct {
	contractName       string
	crdtOperationsList *protos.CRDTOperationsList
}

func NewCRDTOperationsPut(contractName string, crdtOperationsList *protos.CRDTOperationsList) *CRDTOperationsPut {
	contractName = strings.ToLower(contractName)
	return &CRDTOperationsPut{
		contractName:       contractName,
		crdtOperationsList: crdtOperationsList,
	}
}

type CRDTObjectQuery struct {
	contractName string
	crdtObjId    string
}

func NewCRDTObjQueryWithoutChannel(contractName string, crdtObjId string) *CRDTObjectQuery {
	contractName = strings.ToLower(contractName)
	return &CRDTObjectQuery{
		contractName: contractName,
		crdtObjId:    crdtObjId,
	}
}

func (m *Manager) makeNewCRDTNode(clock string) *protos.CRDTObjectNode {
	operationClock, err := StringToClock(clock)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return nil
	}
	return &protos.CRDTObjectNode{
		Map:            map[string]*protos.CRDTObjectNode{},
		LastOperations: []*protos.CRDTClock{operationClock},
	}
}

func (m *Manager) makeNewCRDTNodeWithoutClock() *protos.CRDTObjectNode {
	return &protos.CRDTObjectNode{
		Map:            map[string]*protos.CRDTObjectNode{},
		LastOperations: []*protos.CRDTClock{},
	}
}

func (m *Manager) newCRDTObject(objectId string) *CRDTObject {
	tempCrdtObj := &protos.CRDTObject{
		CrdtObjectId: objectId,
		Head:         m.makeNewCRDTNodeWithoutClock(),
	}
	return &CRDTObject{
		CrdtObj: tempCrdtObj,
		Lock:    &sync.RWMutex{},
	}
}
