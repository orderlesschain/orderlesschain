package contract

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv1"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor/transactiondb"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
)

type Shim struct {
	proposalRequest  *protos.ProposalRequest
	proposalResponse *protos.ProposalResponse
	sharedResources  *contractinterface.SharedShimResources
	dbOp             *transactiondb.Operations
	contractName     string
}

func NewShim(proposal *protos.ProposalRequest, sharedResources *contractinterface.SharedShimResources, contractName string) *Shim {
	return &Shim{
		proposalRequest: proposal,
		proposalResponse: &protos.ProposalResponse{
			ProposalId: proposal.ProposalId,
			NodeId:     config.Config.UUID,
			ReadWriteSet: &protos.ProposalResponseReadWriteSet{
				ReadKeys:       &protos.ProposalResponseReadKeys{ReadKeys: []*protos.ReadKey{}},
				WriteKeyValues: &protos.ProposalResponseWriteKeyValues{WriteKeyValues: []*protos.WriteKeyValue{}},
			},
			NodeSignature: []byte{},
		},
		sharedResources: sharedResources,
		dbOp:            sharedResources.DBConnections[contractName],
		contractName:    contractName,
	}
}

func (s *Shim) GetSharedShimResources() *contractinterface.SharedShimResources {
	return s.sharedResources
}

func (s *Shim) GetKeyValueWithVersion(key string) ([]byte, error) {
	keyValueORM, err := s.dbOp.GetKeyValueWithVersion(key)
	if err != nil {
		return nil, err
	}
	s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
		Key:           key,
		VersionNumber: keyValueORM.VersionNumber,
	})
	return keyValueORM.Value, nil
}

func (s *Shim) GetKeyValueWithVersionZero(key string) ([]byte, error) {
	keyValueORM, err := s.dbOp.GetKeyValueWithVersion(key)
	if err != nil {
		return nil, err
	}
	s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
		Key:           key,
		VersionNumber: 0,
	})
	return keyValueORM.Value, nil
}

func (s *Shim) GetKeyValueNoVersion(key string) ([]byte, error) {
	keyValueByte, err := s.dbOp.GetKeyValueNoVersion(key)
	if err != nil {
		return nil, err
	}
	s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
		Key:           key,
		VersionNumber: 0,
	})
	return keyValueByte, nil
}

func (s *Shim) GetKeyRangeValueWithVersion(key string) (map[string][]byte, error) {
	keyValues := map[string][]byte{}
	keyValueORMs, err := s.dbOp.GetKeyRangeValueWithVersion(key)
	if err != nil {
		return nil, err
	}
	for keyValueORMKey, keyValueORM := range keyValueORMs {
		s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
			Key:           keyValueORMKey,
			VersionNumber: keyValueORM.VersionNumber,
		})
		keyValues[keyValueORMKey] = keyValueORM.Value
	}
	return keyValues, nil
}

func (s *Shim) GetKeyRangeValueWithVersionZero(key string) (map[string][]byte, error) {
	keyValues := map[string][]byte{}
	keyValueORMs, err := s.dbOp.GetKeyRangeValueWithVersion(key)
	if err != nil {
		return nil, err
	}
	for keyValueORMKey, keyValueORM := range keyValueORMs {
		s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
			Key:           keyValueORMKey,
			VersionNumber: 0,
		})
		keyValues[keyValueORMKey] = keyValueORM.Value
	}
	return keyValues, nil
}

func (s *Shim) GetKeyRangeValueNoVersion(key string) (map[string][]byte, error) {
	keyValueORMs, err := s.dbOp.GetKeyRangeNoVersion(key)
	if err != nil {
		return nil, err
	}
	s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
		Key:           key,
		VersionNumber: 0,
	})
	return keyValueORMs, nil
}

func (s *Shim) GetJSONCRDTV1Warm(key string) (*crdtmanagerv1.AuctionContainer, *crdtmanagerv1.PartyContainer, error) {
	auction, party, err := s.sharedResources.CRDTManagerFabricCRDT.GetJSONCRDTValueWarm(s.contractName, key)
	if err != nil {
		return nil, nil, err
	}
	s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
		Key:           key,
		VersionNumber: 0,
	})
	return auction, party, nil
}

func (s *Shim) GetCRDTObjectV2Warm(key string) (*crdtmanagerv2.CRDTObject, error) {
	crdtObjQuery := crdtmanagerv2.NewCRDTObjQueryWithoutChannel(s.contractName, key)
	tempCRDTObj := s.sharedResources.CRDTManager.GetCRDTValueWarm(crdtObjQuery)
	if tempCRDTObj == nil {
		return nil, errors.New("crdt object not found")
	}
	s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
		Key:           key,
		VersionNumber: 0,
	})
	return tempCRDTObj, nil
}

func (s *Shim) GetCRDTObjectV2Cold(key string) (*protos.CRDTObject, error) {
	crdtObjQuery := crdtmanagerv2.NewCRDTObjQueryWithoutChannel(s.contractName, key)
	crdtObject := s.sharedResources.CRDTManager.GetCRDTValueCold(crdtObjQuery)
	return crdtObject, nil
}

func (s *Shim) PutKeyValue(key string, value []byte) {
	s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues = append(s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues, &protos.WriteKeyValue{
		Key:       key,
		Value:     value,
		WriteType: protos.WriteKeyValue_BINARY,
	})
}

func (s *Shim) PutKeyValueFabricCRDT(key string, value []byte) {
	s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys = append(s.proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, &protos.ReadKey{
		Key:           key,
		VersionNumber: 0,
	})
	s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues = append(s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues, &protos.WriteKeyValue{
		Key:       key,
		Value:     value,
		WriteType: protos.WriteKeyValue_FABRICCRDT,
	})
}

func (s *Shim) PutCRDTOperationsV2Warm(key string, crdtOperationsList *protos.CRDTOperationsList) {
	crdtOperationsListByte, _ := proto.Marshal(crdtOperationsList)
	s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues = append(s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues, &protos.WriteKeyValue{
		Key:       key,
		Value:     crdtOperationsListByte,
		WriteType: protos.WriteKeyValue_CRDTOPERATIONSLIST_WARM,
	})
}

func (s *Shim) PutCRDTOperationsV2Cold(key string, crdtOperationsList *protos.CRDTOperationsList) {
	crdtOperationsListByte, _ := proto.Marshal(crdtOperationsList)
	s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues = append(s.proposalResponse.ReadWriteSet.WriteKeyValues.WriteKeyValues, &protos.WriteKeyValue{
		Key:       key,
		Value:     crdtOperationsListByte,
		WriteType: protos.WriteKeyValue_CRDTOPERATIONSLIST_COLD,
	})
}

func (s *Shim) SuccessWithOutput(output []byte) *protos.ProposalResponse {
	s.proposalResponse.Status = protos.ProposalResponse_SUCCESS
	s.proposalResponse.ShimOutput = output
	return s.proposalResponse
}

func (s *Shim) Success() *protos.ProposalResponse {
	s.proposalResponse.Status = protos.ProposalResponse_SUCCESS
	return s.proposalResponse
}

func (s *Shim) Error() *protos.ProposalResponse {
	s.proposalResponse.Status = protos.ProposalResponse_FAIL
	return s.proposalResponse
}
