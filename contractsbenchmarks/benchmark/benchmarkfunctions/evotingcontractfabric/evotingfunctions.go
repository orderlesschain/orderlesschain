package evotingcontractfabric

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
)

func InitElection(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.Counter-1]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId},
		ContractMethodName: "initelection",
	}
}

func Vote(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(AllPartiesCount)]
	voterId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndex()]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId, voterId},
		ContractMethodName: "vote",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func VoteGaussian(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(AllPartiesCount)]
	voterId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndex()]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId, voterId},
		ContractMethodName: "vote",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func QueryParties(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(AllPartiesCount)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId},
		ContractMethodName: "queryparties",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func QueryPartiesGaussian(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(AllPartiesCount)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId},
		ContractMethodName: "queryparties",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func GetElectionResults(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId},
		ContractMethodName: "getelectionresults",
	}
}

func ReadWriteUniformTransaction(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return QueryParties(options)
	} else {
		return Vote(options)
	}
}

func ReadWriteGaussianTransaction(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return QueryPartiesGaussian(options)
	} else {
		return VoteGaussian(options)
	}
}
