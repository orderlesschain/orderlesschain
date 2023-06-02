package contractinterface

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv1"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor/transactiondb"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
)

type SharedShimResources struct {
	DBConnections         map[string]*transactiondb.Operations
	CRDTManager           *crdtmanagerv2.Manager
	CRDTManagerFabricCRDT *crdtmanagerv1.Manager
}

type ShimInterface interface {
	GetKeyValueWithVersion(string) ([]byte, error)
	GetKeyValueWithVersionZero(string) ([]byte, error)
	GetKeyValueNoVersion(string) ([]byte, error)
	GetJSONCRDTV1Warm(string) (*crdtmanagerv1.AuctionContainer, *crdtmanagerv1.PartyContainer, error)
	GetKeyRangeValueWithVersion(string) (map[string][]byte, error)
	GetKeyRangeValueWithVersionZero(string) (map[string][]byte, error)
	GetKeyRangeValueNoVersion(string) (map[string][]byte, error)
	PutKeyValue(string, []byte)
	PutKeyValueFabricCRDT(string, []byte)
	PutCRDTOperationsV2Warm(string, *protos.CRDTOperationsList)
	PutCRDTOperationsV2Cold(string, *protos.CRDTOperationsList)
	GetCRDTObjectV2Warm(string) (*crdtmanagerv2.CRDTObject, error)
	GetCRDTObjectV2Cold(string) (*protos.CRDTObject, error)
	GetSharedShimResources() *SharedShimResources
	SuccessWithOutput([]byte) *protos.ProposalResponse
	Success() *protos.ProposalResponse
	Error() *protos.ProposalResponse
}
