package iotcontractorderlesschain

import (
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strings"
)

type IoTContract struct {
}

func NewContract() *IoTContract {
	return &IoTContract{}
}

func (t *IoTContract) Invoke(shim contractinterface.ShimInterface, proposal *protos.ProposalRequest) (*protos.ProposalResponse, error) {
	proposal.MethodName = strings.ToLower(proposal.MethodName)
	if proposal.MethodName == "submitreading" {
		return t.submitReading(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "monitor" {
		return t.monitor(shim, proposal.MethodParams)
	}
	return shim.Error(), errors.New("method name not found")
}

func (t *IoTContract) submitReading(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	deviceId := args[0]
	containerId := args[1]
	reading := args[2]
	deviceClock := args[3]

	operationsList := &protos.CRDTOperationsList{
		CrdtObjectId: containerId,
		Operations:   []*protos.CRDTOperation{},
	}
	operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
		OperationId:   deviceClock,
		ValueType:     protos.CRDTOperation_MV_REGISTER_STRING,
		Value:         reading,
		OperationPath: []string{deviceId},
		Clock:         deviceClock,
	})
	shim.PutCRDTOperationsV2Warm(containerId, operationsList)

	return shim.Success(), nil
}

func (t *IoTContract) monitor(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	containerId := args[0]

	containerCRDT, errCRDT := shim.GetCRDTObjectV2Warm(containerId)
	if errCRDT != nil {
		return shim.Error(), errCRDT
	}
	containerReadings := map[string]string{}
	containerCRDT.Lock.RLock()
	for readingIndex, reading := range containerCRDT.CrdtObj.Head.Map {
		if reading.MvRegister != nil && len(reading.MvRegister.RegisterValueString) == 1 {
			containerReadings[readingIndex] = reading.MvRegister.RegisterValueString[0]
		}
	}
	containerCRDT.Lock.RUnlock()

	logger.InfoLogger.Println(containerReadings)

	return shim.Success(), nil
}
