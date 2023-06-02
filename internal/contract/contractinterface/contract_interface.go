package contractinterface

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/benchmark/benchmarkutils"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
)

type ContractInterface interface {
	Invoke(ShimInterface, *protos.ProposalRequest) (*protos.ProposalResponse, error)
}

type BaseContractOptions struct {
	Bconfig               *protos.BenchmarkConfig
	BenchmarkFunction     BenchmarkFunction
	BenchmarkUtils        *benchmarkutils.BenchmarkUtils
	CurrentClientPseudoId int
	TotalTransactions     int
	SingleFunctionCounter *SingleFunctionCounter
	NodesConnectionPool   map[string]*connpool.Pool
}
