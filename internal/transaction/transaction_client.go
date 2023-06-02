package transaction

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/signer"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sort"
)

var byzantineTemperedMessage = []byte("tempered")

type ClientTransaction struct {
	TransactionId     string
	ProposalResponses map[string]*protos.ProposalResponse
	Status            protos.TransactionStatus
	signer            *signer.Signer
}

func NewClientTransaction(signer *signer.Signer) *ClientTransaction {
	return &ClientTransaction{
		TransactionId:     uuid.NewString(),
		Status:            protos.TransactionStatus_RUNNING,
		ProposalResponses: map[string]*protos.ProposalResponse{},
		signer:            signer,
	}
}

func (t *ClientTransaction) MakeProposalRequestBenchmarkExecutor(counter int, baseOptions *contractinterface.BaseContractOptions) *protos.ProposalRequest {
	proposalParams := baseOptions.BenchmarkFunction(&contractinterface.BenchmarkFunctionOptions{
		Counter:                     counter,
		CurrentClientPseudoId:       baseOptions.CurrentClientPseudoId,
		BenchmarkUtils:              baseOptions.BenchmarkUtils,
		TotalTransactions:           baseOptions.TotalTransactions,
		SingleFunctionCounter:       baseOptions.SingleFunctionCounter,
		CrdtOperationPerObjectCount: baseOptions.Bconfig.CrdtOperationPerObjectCount,
		CrdtObjectType:              baseOptions.Bconfig.CrdtObjectType,
		CrdtObjectCount:             baseOptions.Bconfig.CrdtObjectCount,
		NumberOfKeys:                int(baseOptions.Bconfig.NumberOfKeys),
		NumberOfKeysSecond:          int(baseOptions.Bconfig.NumberOfKeysSecond),
		NodesConnPool:               baseOptions.NodesConnectionPool,
	})
	return &protos.ProposalRequest{
		TargetSystem:         baseOptions.Bconfig.TargetSystem,
		ProposalId:           t.TransactionId,
		ClientId:             config.Config.UUID,
		ContractName:         baseOptions.Bconfig.ContractName,
		MethodName:           proposalParams.ContractMethodName,
		MethodParams:         proposalParams.Outputs,
		WriteReadTransaction: proposalParams.WriteReadType,
	}
}

func (t *ClientTransaction) MakeTransactionBenchmarkExecutorWithClientFullSign(bconfig *protos.BenchmarkConfig, endorsementPolicy int) (*protos.Transaction, error) {
	if t.Status == protos.TransactionStatus_FAILED_GENERAL {
		return nil, errors.New("transaction failed")
	}
	tempTransaction := &protos.Transaction{
		TargetSystem:      bconfig.TargetSystem,
		TransactionId:     t.TransactionId,
		ClientId:          config.Config.UUID,
		ContractName:      bconfig.ContractName,
		NodeSignatures:    map[string][]byte{},
		EndorsementPolicy: int32(endorsementPolicy),
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.NodeSignatures[proposal.NodeId] = proposal.NodeSignature
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.ReadWriteSet = proposal.ReadWriteSet
		break
	}
	sort.Slice(tempTransaction.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	marshalledReadWriteSet, _ := proto.Marshal(tempTransaction.ReadWriteSet)
	tempTransaction.ClientSignature = t.signer.Sign(marshalledReadWriteSet)

	return tempTransaction, nil
}

func (t *ClientTransaction) MakeTransactionBenchmarkExecutorWithClientBaseSign(bconfig *protos.BenchmarkConfig, endorsementPolicy int) (*protos.Transaction, error) {
	if t.Status == protos.TransactionStatus_FAILED_GENERAL {
		return nil, errors.New("transaction failed")
	}
	tempTransaction := &protos.Transaction{
		TargetSystem:      bconfig.TargetSystem,
		TransactionId:     t.TransactionId,
		ClientId:          config.Config.UUID,
		ContractName:      bconfig.ContractName,
		NodeSignatures:    map[string][]byte{},
		EndorsementPolicy: int32(endorsementPolicy),
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.NodeSignatures[proposal.NodeId] = proposal.NodeSignature
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.ReadWriteSet = proposal.ReadWriteSet
		break
	}
	sort.Slice(tempTransaction.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	marshalledReadSet, _ := proto.Marshal(tempTransaction.ReadWriteSet.ReadKeys)
	tempTransaction.ClientSignature = t.signer.Sign(marshalledReadSet)

	return tempTransaction, nil
}

func (t *ClientTransaction) MakeByzantineTransactionBenchmarkExecutorWithClientFullSign(bconfig *protos.BenchmarkConfig, endorsementPolicy int) (*protos.Transaction, error) {
	if t.Status == protos.TransactionStatus_FAILED_GENERAL {
		return nil, errors.New("transaction failed")
	}
	tempTransaction := &protos.Transaction{
		TargetSystem:      bconfig.TargetSystem,
		TransactionId:     t.TransactionId,
		ClientId:          config.Config.UUID,
		ContractName:      bconfig.ContractName,
		NodeSignatures:    map[string][]byte{},
		EndorsementPolicy: int32(endorsementPolicy),
	}
	for _, proposal := range t.ProposalResponses {
		tempTransaction.NodeSignatures[proposal.NodeId] = proposal.NodeSignature
	}
	for _, proposal := range t.ProposalResponses {
		for _, write := range proposal.ReadWriteSet.WriteKeyValues.WriteKeyValues {
			write.Value = byzantineTemperedMessage
			break
		}
		tempTransaction.ReadWriteSet = proposal.ReadWriteSet
		break
	}
	sort.Slice(tempTransaction.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tempTransaction.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	sort.Slice(tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues, func(i, j int) bool {
		return tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[i].Key < tempTransaction.ReadWriteSet.WriteKeyValues.WriteKeyValues[j].Key
	})
	marshalledReadWriteSet, _ := proto.Marshal(tempTransaction.ReadWriteSet)
	tempTransaction.ClientSignature = t.signer.Sign(marshalledReadWriteSet)

	return tempTransaction, nil
}

func MakeQueryProposalRequest(targetName protos.TargetSystem, contractName string, methodName string, methodParams []string) *protos.ProposalRequest {
	ProposalRequest := &protos.ProposalRequest{
		TargetSystem: targetName,
		ProposalId:   uuid.NewString(),
		ClientId:     config.Config.UUID,
		ContractName: contractName,
		MethodName:   methodName,
		MethodParams: methodParams,
	}
	return ProposalRequest
}

func (t *ClientTransaction) MakeProposalBIDLAndSyncHotStuffClientSign(proposalRequest *protos.ProposalRequest) (*protos.ProposalRequest, error) {
	marshalledProposal, _ := proto.Marshal(&protos.ProposalRequest{
		ProposalId:   proposalRequest.ProposalId,
		ClientId:     proposalRequest.ClientId,
		ContractName: proposalRequest.ContractName,
		MethodName:   proposalRequest.MethodName,
		MethodParams: proposalRequest.MethodParams,
	})
	proposalRequest.ClientSignature = t.signer.Sign(marshalledProposal)
	return proposalRequest, nil
}
