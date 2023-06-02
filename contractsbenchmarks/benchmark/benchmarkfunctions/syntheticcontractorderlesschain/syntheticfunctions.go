package syntheticcontractorderlesschain

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
)

func ModifyCRDTWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	userIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	userId := UserIDs[userIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[userIdIndex]++
	counter := options.SingleFunctionCounter.Counter[userIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(userId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs: []string{userId + "#", options.CrdtObjectCount, options.CrdtOperationPerObjectCount,
			options.CrdtObjectType, crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "modifyCRDTWarm",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func ModifyCRDTCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	userIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	userId := UserIDs[userIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[userIdIndex]++
	counter := options.SingleFunctionCounter.Counter[userIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(userId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs: []string{userId + "#", options.CrdtObjectCount, options.CrdtOperationPerObjectCount,
			options.CrdtObjectType, crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "modifyCRDTCold",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func ReadCRDTWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{options.CrdtObjectCount},
		ContractMethodName: "readCRDTWarm",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func ReadCRDTCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{options.CrdtObjectCount},
		ContractMethodName: "readCRDTCold",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func ReadWriteUniformTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read10Write90TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read10Write90() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read20Write80TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read20Write80() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read30Write70TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read30Write70() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read40Write60TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read40Write60() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read50Write50TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read60Write40TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read60Write40() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read70Write30TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read70Write30() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read80Write20TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read80Write20() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func Read90Write10TransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read90Write10() {
		return ReadCRDTWarm(options)
	} else {
		return ModifyCRDTWarm(options)
	}
}

func ReadWriteUniformTransactionCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return ReadCRDTCold(options)
	} else {
		return ModifyCRDTCold(options)
	}
}
