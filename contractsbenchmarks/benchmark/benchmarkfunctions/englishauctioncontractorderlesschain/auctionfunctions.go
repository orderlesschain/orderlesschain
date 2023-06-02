package englishauctioncontractorderlesschain

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
)

func OfferBidWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	bidderIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	bidderId := AllBidderAuctionsIds[bidderIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[bidderIdIndex]++
	counter := options.SingleFunctionCounter.Counter[bidderIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(bidderId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId, bidderId + "#", strconv.Itoa(options.Counter), crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "offerbidwarm",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func OfferBidGaussianWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	bidderIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	bidderId := AllBidderAuctionsIds[bidderIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[bidderIdIndex]++
	counter := options.SingleFunctionCounter.Counter[bidderIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(bidderId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId, bidderId + "#", strconv.Itoa(options.Counter), crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "offerbidwarm",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func OfferBidCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	bidderIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	bidderId := AllBidderAuctionsIds[bidderIdIndex]
	options.SingleFunctionCounter.Lock.Lock()
	options.SingleFunctionCounter.Counter[bidderIdIndex]++
	counter := options.SingleFunctionCounter.Counter[bidderIdIndex]
	options.SingleFunctionCounter.Lock.Unlock()
	clock := crdtmanagerv2.NewClockWithCounter(bidderId, int64(counter+(options.CurrentClientPseudoId*options.TotalTransactions)))
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId, bidderId + "#", strconv.Itoa(options.Counter), crdtmanagerv2.ClockToString(clock)},
		ContractMethodName: "offerbidcold",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func GetCurrentHighestPriceWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getcurrenthighestpricewarm",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func GetCurrentHighestPriceGaussianWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getcurrenthighestpricewarm",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func GetCurrentHighestPriceCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getcurrenthighestpricecold",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func GetAuctionWinnerWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getauctionwinnerwarm",
	}
}

func GetAuctionWinnerCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getauctionwinnercold",
	}
}

func ReadWriteUniformTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return GetCurrentHighestPriceWarm(options)
	} else {
		return OfferBidWarm(options)
	}
}

func ReadWriteGaussianTransactionWarm(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return GetCurrentHighestPriceGaussianWarm(options)
	} else {
		return OfferBidGaussianWarm(options)
	}
}

func ReadWriteUniformTransactionCold(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return GetCurrentHighestPriceCold(options)
	} else {
		return OfferBidCold(options)
	}
}
