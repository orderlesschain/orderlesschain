package benchmarkfunctions

import (
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/englishauctioncontractbidl"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/englishauctioncontractfabric"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/englishauctioncontractfabriccrdt"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/englishauctioncontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/englishauctioncontractsynchotstuff"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/evotingcontractbidl"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/evotingcontractfabric"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/evotingcontractfabriccrdt"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/evotingcontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/evotingcontractsynchotstuff"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/iotcontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/contractsbenchmarks/benchmark/benchmarkfunctions/syntheticcontractorderlesschain"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"strings"
)

func GetBenchmarkFunctions(contactName string, benchmarkFunctionName string) (contractinterface.BenchmarkFunction, error) {
	contactName = strings.ToLower(contactName)
	benchmarkFunctionName = strings.ToLower(benchmarkFunctionName)
	switch contactName {
	case "syntheticcontractorderlesschain":
		switch benchmarkFunctionName {
		case "readwritetransactionwarm":
			return syntheticcontractorderlesschain.ReadWriteUniformTransactionWarm, nil
		case "read10write90transactionwarm":
			return syntheticcontractorderlesschain.Read10Write90TransactionWarm, nil
		case "read20write80transactionwarm":
			return syntheticcontractorderlesschain.Read20Write80TransactionWarm, nil
		case "read30write70transactionwarm":
			return syntheticcontractorderlesschain.Read30Write70TransactionWarm, nil
		case "read40write60transactionwarm":
			return syntheticcontractorderlesschain.Read40Write60TransactionWarm, nil
		case "read50write50transactionwarm":
			return syntheticcontractorderlesschain.Read50Write50TransactionWarm, nil
		case "read60write40transactionwarm":
			return syntheticcontractorderlesschain.Read60Write40TransactionWarm, nil
		case "read70write30transactionwarm":
			return syntheticcontractorderlesschain.Read70Write30TransactionWarm, nil
		case "read80write20transactionwarm":
			return syntheticcontractorderlesschain.Read80Write20TransactionWarm, nil
		case "read90write10transactionwarm":
			return syntheticcontractorderlesschain.Read90Write10TransactionWarm, nil
		case "readwritetransactioncold":
			return syntheticcontractorderlesschain.ReadWriteUniformTransactionCold, nil
		}
	case "evotingcontractorderlesschain":
		switch benchmarkFunctionName {
		case "getelectionresultswarm":
			return evotingcontractorderlesschain.GetElectionResultsWarm, nil
		case "getelectionresultscold":
			return evotingcontractorderlesschain.GetElectionResultsCold, nil
		case "readwritetransactionwarm":
			return evotingcontractorderlesschain.ReadWriteUniformTransactionWarm, nil
		case "readwritegaussian":
			return evotingcontractorderlesschain.ReadWriteGaussianTransactionWarm, nil
		case "readwritetransactioncold":
			return evotingcontractorderlesschain.ReadWriteUniformTransactionCold, nil
		}
	case "evotingcontractbidl":
		switch benchmarkFunctionName {
		case "initelection":
			return evotingcontractbidl.InitElection, nil
		case "getelectionresults":
			return evotingcontractbidl.GetElectionResults, nil
		case "readwritetransaction":
			return evotingcontractbidl.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return evotingcontractbidl.ReadWriteGaussianTransaction, nil
		}
	case "evotingcontractsynchotstuff":
		switch benchmarkFunctionName {
		case "initelection":
			return evotingcontractsynchotstuff.InitElection, nil
		case "getelectionresults":
			return evotingcontractsynchotstuff.GetElectionResults, nil
		case "readwritetransaction":
			return evotingcontractsynchotstuff.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return evotingcontractsynchotstuff.ReadWriteGaussianTransaction, nil
		}
	case "evotingcontractfabric":
		switch benchmarkFunctionName {
		case "initelection":
			return evotingcontractfabric.InitElection, nil
		case "getelectionresults":
			return evotingcontractfabric.GetElectionResults, nil
		case "readwritetransaction":
			return evotingcontractfabric.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return evotingcontractfabric.ReadWriteGaussianTransaction, nil
		}
	case "evotingcontractfabriccrdt":
		switch benchmarkFunctionName {
		case "initelection":
			return evotingcontractfabriccrdt.InitElection, nil
		case "getelectionresults":
			return evotingcontractfabriccrdt.GetElectionResults, nil
		case "readwritetransaction":
			return evotingcontractfabriccrdt.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return evotingcontractfabriccrdt.ReadWriteGaussianTransaction, nil
		}
	case "englishauctioncontractorderlesschain":
		switch benchmarkFunctionName {
		case "getauctionwinnerwarm":
			return englishauctioncontractorderlesschain.GetAuctionWinnerWarm, nil
		case "getauctionwinnercold":
			return englishauctioncontractorderlesschain.GetAuctionWinnerCold, nil
		case "readwritetransactionwarm":
			return englishauctioncontractorderlesschain.ReadWriteUniformTransactionWarm, nil
		case "readwritegaussian":
			return englishauctioncontractorderlesschain.ReadWriteGaussianTransactionWarm, nil
		case "readwritetransactioncold":
			return englishauctioncontractorderlesschain.ReadWriteUniformTransactionCold, nil
		}
	case "englishauctioncontractfabric":
		switch benchmarkFunctionName {
		case "createauction":
			return englishauctioncontractfabric.CreateAuction, nil
		case "getauctionwinner":
			return englishauctioncontractfabric.GetAuctionWinner, nil
		case "readwritetransaction":
			return englishauctioncontractfabric.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return englishauctioncontractfabric.ReadWriteGaussianTransaction, nil
		}
	case "englishauctioncontractbidl":
		switch benchmarkFunctionName {
		case "createauction":
			return englishauctioncontractbidl.CreateAuction, nil
		case "getauctionwinner":
			return englishauctioncontractbidl.GetAuctionWinner, nil
		case "readwritetransaction":
			return englishauctioncontractbidl.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return englishauctioncontractbidl.ReadWriteGaussianTransaction, nil
		}
	case "englishauctioncontractsynchotstuff":
		switch benchmarkFunctionName {
		case "createauction":
			return englishauctioncontractsynchotstuff.CreateAuction, nil
		case "getauctionwinner":
			return englishauctioncontractsynchotstuff.GetAuctionWinner, nil
		case "readwritetransaction":
			return englishauctioncontractsynchotstuff.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return englishauctioncontractsynchotstuff.ReadWriteGaussianTransaction, nil
		}
	case "englishauctioncontractfabriccrdt":
		switch benchmarkFunctionName {
		case "createauction":
			return englishauctioncontractfabriccrdt.CreateAuction, nil
		case "getauctionwinner":
			return englishauctioncontractfabriccrdt.GetAuctionWinner, nil
		case "readwritetransaction":
			return englishauctioncontractfabriccrdt.ReadWriteUniformTransaction, nil
		case "readwritegaussian":
			return englishauctioncontractfabriccrdt.ReadWriteGaussianTransaction, nil
		}
	case "iotcontractorderlesschain":
		switch benchmarkFunctionName {
		case "submitreading":
			return iotcontractorderlesschain.SubmitReading, nil
		case "monitor":
			return iotcontractorderlesschain.Monitor, nil
		case "readwritetransactionwarm":
			return iotcontractorderlesschain.ReadWriteUniformTransactionWarm, nil
		}
	}
	logger.ErrorLogger.Println("smart contract function was not found")
	return nil, errors.New("functions not found")
}
