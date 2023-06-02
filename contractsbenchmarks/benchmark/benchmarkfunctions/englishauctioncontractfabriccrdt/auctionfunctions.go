package englishauctioncontractfabriccrdt

import (
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
)

func CreateAuction(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.Counter-1]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "createauction",
	}
}

func OfferBid(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	bidderIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	bidderId := AllBidderAuctionsIds[bidderIdIndex]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId, bidderId + "#", strconv.Itoa(options.Counter)},
		ContractMethodName: "offerbid",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func OfferBidGaussian(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	bidderIdIndex := options.BenchmarkUtils.GetUniformDistributionKeyIndex()
	bidderId := AllBidderAuctionsIds[bidderIdIndex]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId, bidderId + "#", strconv.Itoa(options.Counter)},
		ContractMethodName: "offerbid",
		WriteReadType:      protos.ProposalRequest_Write,
	}
}

func GetCurrentHighestPrice(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getcurrenthighestprice",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func GetCurrentHighestPriceGaussian(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getcurrenthighestprice",
		WriteReadType:      protos.ProposalRequest_Read,
	}
}

func GetAuctionWinner(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	auctionId := AllBidderAuctionsIds[options.BenchmarkUtils.GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(options.NumberOfKeysSecond)]
	return &contractinterface.BenchmarkFunctionOutputs{
		Outputs:            []string{auctionId},
		ContractMethodName: "getauctionwinner",
	}
}

func ReadWriteUniformTransaction(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return GetCurrentHighestPrice(options)
	} else {
		return OfferBid(options)
	}
}

func ReadWriteGaussianTransaction(options *contractinterface.BenchmarkFunctionOptions) *contractinterface.BenchmarkFunctionOutputs {
	if options.BenchmarkUtils.Read50Write50() {
		return GetCurrentHighestPriceGaussian(options)
	} else {
		return OfferBidGaussian(options)
	}
}
