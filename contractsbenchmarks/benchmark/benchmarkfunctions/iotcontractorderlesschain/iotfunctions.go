package iotcontractorderlesschain

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
)

func SubmitReading(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	deviceId := AllIoTDevices[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	containerIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	containerId := AllContainerIDs[containerIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[containerIndex]++
	counter := options.SingleFunctionCounter.Counter[containerIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(containerId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	reading := strconv.Itoa(int(options.BenchmarkUtils.GetRandomFloatWithMinAndMax(MinimumTemp, MaximumTemp)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{deviceId, containerId, reading, crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "submitreading",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func Monitor(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	containerIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	containerId := AllContainerIDs[containerIndex]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{containerId},
		ContractMethodName: "monitor",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func ReadWriteUniformTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return SubmitReading(options)
	} else {
		return Monitor(options)
	}
}
