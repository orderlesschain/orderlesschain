package contracts

import (
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/englishauctioncontractbidl"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/englishauctioncontractfabric"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/englishauctioncontractfabriccrdt"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/englishauctioncontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/englishauctioncontractsynchotstuff"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/evotingcontractbidl"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/evotingcontractfabric"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/evotingcontractfabriccrdt"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/evotingcontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/evotingcontractsynchotstuff"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/iotcontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/contracts/syntheticcontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"strings"
)

func GetContract(contractName string) (contractinterface.ContractInterface, error) {
	contractName = strings.ToLower(contractName)
	switch contractName {
	case "evotingcontractbidl":
		return evotingcontractbidl.NewContract(), nil
	case "evotingcontractsynchotstuff":
		return evotingcontractsynchotstuff.NewContract(), nil
	case "evotingcontractfabric":
		return evotingcontractfabric.NewContract(), nil
	case "evotingcontractorderlesschain":
		return evotingcontractorderlesschain.NewContract(), nil
	case "evotingcontractfabriccrdt":
		return evotingcontractfabriccrdt.NewContract(), nil
	case "englishauctioncontractfabric":
		return englishauctioncontractfabric.NewContract(), nil
	case "englishauctioncontractbidl":
		return englishauctioncontractbidl.NewContract(), nil
	case "englishauctioncontractsynchotstuff":
		return englishauctioncontractsynchotstuff.NewContract(), nil
	case "englishauctioncontractorderlesschain":
		return englishauctioncontractorderlesschain.NewContract(), nil
	case "englishauctioncontractfabriccrdt":
		return englishauctioncontractfabriccrdt.NewContract(), nil
	case "syntheticcontractorderlesschain":
		return syntheticcontractorderlesschain.NewContract(), nil
	case "iotcontractorderlesschain":
		return iotcontractorderlesschain.NewContract(), nil
	default:
		return nil, errors.New("contract not found")
	}
}

func GetContractNames() []string {
	return []string{
		"englishauctioncontractbidl",
		"englishauctioncontractsynchotstuff",
		"englishauctioncontractorderlesschain",
		"englishauctioncontractfabric",
		"englishauctioncontractfabriccrdt",
		"evotingcontractfabric",
		"evotingcontractfabriccrdt",
		"evotingcontractbidl",
		"evotingcontractsynchotstuff",
		"evotingcontractorderlesschain",
		"syntheticcontractorderlesschain",
		"filestoragecontractorderlesschain",
		"iotcontractorderlesschain",
	}
}
