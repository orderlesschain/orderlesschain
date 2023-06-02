package syntheticcontractorderlesschain

import (
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
	"strings"
)

const CRDTObjectBaseID = "base"

type SyntheticContract struct {
}

func NewContract() *SyntheticContract {
	return &SyntheticContract{}
}

func (e *SyntheticContract) Invoke(shim contractinterface.ShimInterface, proposal *protos.ProposalRequest) (*protos.ProposalResponse, error) {
	proposal.MethodName = strings.ToLower(proposal.MethodName)

	if proposal.MethodName == "modifycrdtwarm" {
		return e.modifyCRDTWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "modifycrdtcold" {
		return e.modifyCRDTCold(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "readcrdtwarm" {
		return e.readCRDTWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "readcrdtcold" {
		return e.readCRDTCold(shim, proposal.MethodParams)
	}

	return shim.Error(), errors.New("method name not found")
}

func (e *SyntheticContract) modifyCRDTWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	userIdString := args[0]
	crdtObjectCountString := args[1]
	operationPerObjectCountString := args[2]
	crdtTypeString := args[3]
	clockString := args[4]

	crdtObjectCount, err := strconv.Atoi(crdtObjectCountString)
	if err != nil {
		return shim.Error(), err
	}

	operationPerObjectCount, err := strconv.Atoi(operationPerObjectCountString)
	if err != nil {
		return shim.Error(), err
	}

	var crdtType protos.CRDTOperation_ValueType
	var crdtValue string

	if crdtTypeString == "map" {
		crdtType = protos.CRDTOperation_MAP
		crdtValue = "mapValue"
	} else if crdtTypeString == "register" {
		crdtType = protos.CRDTOperation_MV_REGISTER
		crdtValue = "true"
	} else if crdtTypeString == "counter" {
		crdtType = protos.CRDTOperation_G_COUNTER
		crdtValue = "5"
	} else {
		return shim.Error(), err
	}

	for i := 0; i < crdtObjectCount; i++ {
		crdtObjectId := CRDTObjectBaseID + strconv.Itoa(i)
		operationsList := &protos.CRDTOperationsList{
			CrdtObjectId: crdtObjectId,
			Operations:   []*protos.CRDTOperation{},
		}
		for j := 0; j < operationPerObjectCount; j++ {
			objectPathPart := "#" + strconv.Itoa(j)
			operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
				OperationId:   clockString + objectPathPart,
				ValueType:     crdtType,
				Value:         crdtValue,
				OperationPath: []string{userIdString, objectPathPart},
				Clock:         clockString,
			})
		}
		shim.PutCRDTOperationsV2Warm(crdtObjectId, operationsList)
	}
	return shim.Success(), nil
}

func (e *SyntheticContract) modifyCRDTCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	userIdString := args[0]
	crdtObjectCountString := args[1]
	operationPerObjectCountString := args[2]
	crdtTypeString := args[3]
	clockString := args[4]

	crdtObjectCount, err := strconv.Atoi(crdtObjectCountString)
	if err != nil {
		return shim.Error(), err
	}

	operationPerObjectCount, err := strconv.Atoi(operationPerObjectCountString)
	if err != nil {
		return shim.Error(), err
	}

	var crdtType protos.CRDTOperation_ValueType
	var crdtValue string

	if crdtTypeString == "map" {
		crdtType = protos.CRDTOperation_MAP
		crdtValue = "mapValue"
	} else if crdtTypeString == "register" {
		crdtType = protos.CRDTOperation_MV_REGISTER
		crdtValue = "true"
	} else if crdtTypeString == "counter" {
		crdtType = protos.CRDTOperation_G_COUNTER
		crdtValue = "5"
	} else {
		return shim.Error(), err
	}

	for i := 0; i < crdtObjectCount; i++ {
		crdtObjectId := CRDTObjectBaseID + strconv.Itoa(i)
		operationsList := &protos.CRDTOperationsList{
			CrdtObjectId: crdtObjectId,
			Operations:   []*protos.CRDTOperation{},
		}
		for j := 0; j < operationPerObjectCount; j++ {
			objectPathPart := "#" + strconv.Itoa(j)
			operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
				OperationId:   clockString + objectPathPart,
				ValueType:     crdtType,
				Value:         crdtValue,
				OperationPath: []string{userIdString, objectPathPart},
				Clock:         clockString,
			})
		}
		shim.PutCRDTOperationsV2Cold(crdtObjectId, operationsList)
	}
	return shim.Success(), nil
}

func (e *SyntheticContract) readCRDTWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	crdtObjectCountString := args[0]
	crdtObjectCount, err := strconv.Atoi(crdtObjectCountString)
	if err != nil {
		return shim.Error(), err
	}

	for i := 0; i < crdtObjectCount; i++ {
		crdtObjectId := CRDTObjectBaseID + strconv.Itoa(i)
		_, errCRDT := shim.GetCRDTObjectV2Warm(crdtObjectId)
		if errCRDT != nil {
			return shim.Error(), err
		}
	}
	return shim.Success(), nil
}

func (e *SyntheticContract) readCRDTCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	crdtObjectCountString := args[0]
	crdtObjectCount, err := strconv.Atoi(crdtObjectCountString)
	if err != nil {
		return shim.Error(), err
	}

	for i := 0; i < crdtObjectCount; i++ {
		crdtObjectId := CRDTObjectBaseID + strconv.Itoa(i)
		_, errCRDT := shim.GetCRDTObjectV2Cold(crdtObjectId)
		if errCRDT != nil {
			return shim.Error(), err
		}
	}
	return shim.Success(), nil
}
