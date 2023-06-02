package evotingcontractorderlesschain

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
)

func VoteWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(AllPartiesCount)]
	voterIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	voterId := AllVoterElectionIDs[voterIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[voterIdIndex]++
	counter := options.SingleFunctionCounter.Counter[voterIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(voterId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId, voterId + "#", crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "votewarm",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func VoteGaussianWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(AllPartiesCount)]
	voterIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	voterId := AllVoterElectionIDs[voterIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[voterIdIndex]++
	counter := options.SingleFunctionCounter.Counter[voterIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(voterId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId, voterId + "#", crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "votewarm",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func VoteCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(AllPartiesCount)]
	voterIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	voterId := AllVoterElectionIDs[voterIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[voterIdIndex]++
	counter := options.SingleFunctionCounter.Counter[voterIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(voterId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId, voterId + "#", crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "votecold",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func QueryPartyVoteCountWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(AllPartiesCount)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId},
		ContractMethodName: "querypartieswarm",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func QueryPartyVoteCountGaussianWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(AllPartiesCount)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId},
		ContractMethodName: "querypartieswarm",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func QueryPartyVoteCountCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	partyId := AllParties[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(AllPartiesCount)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId, partyId},
		ContractMethodName: "querypartiescold",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func GetElectionResultsWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId},
		ContractMethodName: "getelectionresultswarm",
	}
}

func GetElectionResultsCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	electionId := AllVoterElectionIDs[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{electionId},
		ContractMethodName: "getelectionresultscold",
	}
}

func ReadWriteUniformTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return QueryPartyVoteCountWarm(options)
	} else {
		return VoteWarm(options)
	}
}

func ReadWriteGaussianTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return QueryPartyVoteCountGaussianWarm(options)
	} else {
		return VoteGaussianWarm(options)
	}
}

func ReadWriteUniformTransactionCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return QueryPartyVoteCountCold(options)
	} else {
		return VoteCold(options)
	}
}
