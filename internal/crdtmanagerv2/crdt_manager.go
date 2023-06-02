package crdtmanagerv2

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
)

func (m *Manager) createNodesFromRootToLeaf(leafNode *protos.CRDTObjectNode, path []string, clock string) *protos.CRDTObjectNode {
	for _, pathString := range path {
		if _, ok := leafNode.Map[pathString]; !ok {
			leafNode.Map[pathString] = m.makeNewCRDTNode(clock)
		}
		leafNode = leafNode.Map[pathString]
	}
	return leafNode
}

func (m *Manager) modifyGCounter(leafNode *protos.CRDTObjectNode, operation *protos.CRDTOperation) {
	counter := leafNode.GCounter
	if counter == nil {
		counter = &protos.GCounter{}
		leafNode.GCounter = counter
	}
	valueInt, err := strconv.Atoi(operation.Value)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	// No need to check for happened before as it is intrinsically concurrent
	counter.CounterValue += int32(valueInt)
}

func (m *Manager) modifyMVRegister(leafNode *protos.CRDTObjectNode, operation *protos.CRDTOperation) {
	mvRegister := leafNode.MvRegister
	if mvRegister == nil {
		mvRegister = &protos.MVRegister{
			RegisterValue: []bool{},
		}
		leafNode.MvRegister = mvRegister
		leafNode.LastOperations = []*protos.CRDTClock{}
	}
	valueBool, err := strconv.ParseBool(operation.Value)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	operationClock, err := StringToClock(operation.Clock)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	operationsCountInNode := len(leafNode.LastOperations)
	if operationsCountInNode == 0 {
		mvRegister.RegisterValue = append(mvRegister.RegisterValue, valueBool)
		leafNode.LastOperations = append(leafNode.LastOperations, operationClock)
	} else {
		lastClock := operationsCountInNode - 1
		happenedBefore := ClockHappenedBefore(leafNode.LastOperations[lastClock], operationClock)
		if happenedBefore == NoHappenedBefore {
			mvRegister.RegisterValue = append(mvRegister.RegisterValue, valueBool)
			leafNode.LastOperations = append(leafNode.LastOperations, operationClock)
		} else if happenedBefore == FirstHappenedBeforeSecond {
			mvRegister.RegisterValue[len(mvRegister.RegisterValue)-1] = valueBool
			leafNode.LastOperations[lastClock] = operationClock
		}
	}
}

func (m *Manager) modifyMVRegisterString(leafNode *protos.CRDTObjectNode, operation *protos.CRDTOperation) {
	mvRegister := leafNode.MvRegister
	if mvRegister == nil {
		mvRegister = &protos.MVRegister{
			RegisterValueString: []string{},
		}
		leafNode.MvRegister = mvRegister
		leafNode.LastOperations = []*protos.CRDTClock{}
	}
	operationClock, err := StringToClock(operation.Clock)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	operationsCountInNode := len(leafNode.LastOperations)
	if operationsCountInNode == 0 {
		mvRegister.RegisterValueString = append(mvRegister.RegisterValueString, operation.Value)
		leafNode.LastOperations = append(leafNode.LastOperations, operationClock)
	} else {
		lastClock := operationsCountInNode - 1
		happenedBefore := ClockHappenedBefore(leafNode.LastOperations[lastClock], operationClock)
		if happenedBefore == NoHappenedBefore {
			mvRegister.RegisterValueString = append(mvRegister.RegisterValueString, operation.Value)
			leafNode.LastOperations = append(leafNode.LastOperations, operationClock)
		} else if happenedBefore == FirstHappenedBeforeSecond {
			mvRegister.RegisterValueString[len(mvRegister.RegisterValueString)-1] = operation.Value
			leafNode.LastOperations[lastClock] = operationClock
		}
	}
}

func (m *Manager) modifyMap(leafNode *protos.CRDTObjectNode, operation *protos.CRDTOperation) {
	operationsCountInNode := len(leafNode.LastOperations)
	operationClock, err := StringToClock(operation.Clock)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	happenedBefore := ClockHappenedBefore(leafNode.LastOperations[operationsCountInNode-1], operationClock)
	if happenedBefore == NoHappenedBefore {
		lastClockString := ClockToString(leafNode.LastOperations[operationsCountInNode-1])
		thisClockString := ClockToString(operationClock)
		leafNode.Map[lastClockString] = m.makeNewCRDTNode(lastClockString)
		leafNode.Map[lastClockString].MapValue = leafNode.MapValue
		leafNode.Map[thisClockString] = m.makeNewCRDTNode(thisClockString)
		leafNode.Map[thisClockString].MapValue = operation.Value
		leafNode.LastOperations = append(leafNode.LastOperations, operationClock)
	} else if happenedBefore == FirstHappenedBeforeSecond {
		leafNode.LastOperations[operationsCountInNode-1] = operationClock
		leafNode.MapValue = operation.Value
	}
}

func (m *Manager) modifyLeafNode(leafNode *protos.CRDTObjectNode, operation *protos.CRDTOperation) {
	switch operation.ValueType {
	case protos.CRDTOperation_G_COUNTER:
		m.modifyGCounter(leafNode, operation)
	case protos.CRDTOperation_MV_REGISTER:
		m.modifyMVRegister(leafNode, operation)
	case protos.CRDTOperation_MAP:
		m.modifyMap(leafNode, operation)
	case protos.CRDTOperation_MV_REGISTER_STRING:
		m.modifyMVRegisterString(leafNode, operation)
	}
}

func (m *Manager) applyOperationsToObject(crdtObj *CRDTObject, operations []*protos.CRDTOperation) {
	crdtObj.Lock.Lock()
	for _, operation := range operations {
		leafNode := m.createNodesFromRootToLeaf(crdtObj.CrdtObj.Head, operation.OperationPath, operation.Clock)
		m.modifyLeafNode(leafNode, operation)
	}
	crdtObj.Lock.Unlock()
}

func (m *Manager) applyOperationsDBToObject(contractName string, crdtObj *CRDTObject) {
	operationsList, err := m.DBConnections[contractName].GetCRDTOperationsV2(crdtObj.CrdtObj.CrdtObjectId)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	m.applyOperationsToObject(crdtObj, operationsList.Operations)
}

func (m *Manager) ApplyOperationsWarm(crdtOperationsPut *CRDTOperationsPut) {
	crdtObjId := crdtOperationsPut.contractName + crdtOperationsPut.crdtOperationsList.CrdtObjectId
	m.lock.RLock()
	crdtObj, ok := m.crdtObjs[crdtObjId]
	m.lock.RUnlock()
	if !ok {
		crdtObj = m.newCRDTObject(crdtOperationsPut.crdtOperationsList.CrdtObjectId)
		m.lock.Lock()
		m.crdtObjs[crdtObjId] = crdtObj
		m.lock.Unlock()
		// !! This is commented out, as for this evaluation, it does not make any difference, as it is only called once during initialization
		// Applying DB operations could cause, applying the operations multiple times
		//m.applyOperationsDBToObject(crdtOperationsPut.contractName, crdtObj)
	}
	m.applyOperationsToObject(crdtObj, crdtOperationsPut.crdtOperationsList.Operations)
}

func (m *Manager) GetCRDTValueWarm(crdtObjectQuery *CRDTObjectQuery) *CRDTObject {
	crdtObjId := crdtObjectQuery.contractName + crdtObjectQuery.crdtObjId
	m.lock.RLock()
	crdtObj, ok := m.crdtObjs[crdtObjId]
	m.lock.RUnlock()
	if !ok {
		crdtObj = m.newCRDTObject(crdtObjectQuery.crdtObjId)
		m.lock.Lock()
		m.crdtObjs[crdtObjId] = crdtObj
		m.lock.Unlock()
	}
	return crdtObj
}

func (m *Manager) GetCRDTValueCold(crdtObjectQuery *CRDTObjectQuery) *protos.CRDTObject {
	crdtObj := m.newCRDTObject(crdtObjectQuery.crdtObjId)
	m.applyOperationsDBToObject(crdtObjectQuery.contractName, crdtObj)
	return crdtObj.CrdtObj
}
