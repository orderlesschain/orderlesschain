package englishauctioncontractorderlesschain

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv2"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strings"
)

type EnglishAuctionContract struct {
}

func NewContract() *EnglishAuctionContract {
	return &EnglishAuctionContract{}
}

func (e *EnglishAuctionContract) Invoke(shim contractinterface.ShimInterface, proposal *protos.ProposalRequest) (*protos.ProposalResponse, error) {
	proposal.MethodName = strings.ToLower(proposal.MethodName)

	if proposal.MethodName == "offerbidwarm" {
		return e.offerBidWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "offerbidcold" {
		return e.offerBidCold(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getcurrenthighestpricewarm" {
		return e.getCurrentHighestPriceWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getcurrenthighestpricecold" {
		return e.getCurrentHighestPriceCold(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getauctionwinnerwarm" {
		return e.getAuctionWinnerWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getauctionwinnercold" {
		return e.getAuctionWinnerCold(shim, proposal.MethodParams)
	}

	return shim.Error(), errors.New("method name not found")
}

func (e *EnglishAuctionContract) offerBidWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	bidderId := args[1]
	offeredPriceString := args[2]
	bidderClockString := args[3]

	operationsList := &protos.CRDTOperationsList{
		CrdtObjectId: auctionId,
		Operations:   []*protos.CRDTOperation{},
	}
	operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
		OperationId:   bidderClockString,
		OperationPath: []string{bidderId},
		ValueType:     protos.CRDTOperation_G_COUNTER,
		Value:         offeredPriceString,
		Clock:         bidderClockString,
	})
	shim.PutCRDTOperationsV2Warm(auctionId, operationsList)
	return shim.Success(), nil
}

func (e *EnglishAuctionContract) offerBidCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	bidderId := args[1]
	offeredPriceString := args[2]
	bidderClockString := args[3]

	operationsList := &protos.CRDTOperationsList{
		CrdtObjectId: auctionId,
		Operations:   []*protos.CRDTOperation{},
	}
	operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
		OperationId:   bidderClockString,
		OperationPath: []string{bidderId},
		ValueType:     protos.CRDTOperation_G_COUNTER,
		Value:         offeredPriceString,
		Clock:         bidderClockString,
	})
	shim.PutCRDTOperationsV2Cold(auctionId, operationsList)
	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getCurrentHighestPriceWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	auctionCRDT, err := shim.GetCRDTObjectV2Warm(auctionId)
	if err != nil {
		return shim.Error(), err
	}
	bidderId, price, err := e.getHighestPriceFromCRDTWarm(auctionCRDT)
	if err != nil {
		return shim.Error(), err
	}
	_, _ = bidderId, price
	//logger.WarningLogger.Println(bidderId, price)
	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getCurrentHighestPriceCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	auctionCRDT, err := shim.GetCRDTObjectV2Cold(auctionId)
	if err != nil {
		return shim.Error(), err
	}
	bidderId, price, err := e.getHighestPriceFromCRDTCOLD(auctionCRDT)
	if err != nil {
		return shim.Error(), err
	}
	_, _ = bidderId, price
	//logger.WarningLogger.Println(bidderId, price)
	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getAuctionWinnerWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Getting the auction winner warm")
	auctionId := args[0]
	auctionCRDT, err := shim.GetCRDTObjectV2Warm(auctionId)
	if err != nil {
		return shim.Error(), err
	}
	bidderId, price, err := e.getHighestPriceFromCRDTWarm(auctionCRDT)
	if err != nil {
		return shim.Error(), err
	}
	logger.InfoLogger.Println("Highest Price", bidderId, price)
	winner := map[string]int32{bidderId: price}
	winnerBuffed, err := json.Marshal(winner)
	if err != nil {
		return shim.Error(), err
	}
	return shim.SuccessWithOutput(winnerBuffed), nil
}

func (e *EnglishAuctionContract) getAuctionWinnerCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Getting the auction winner cold")
	auctionId := args[0]
	auctionCRDT, err := shim.GetCRDTObjectV2Cold(auctionId)
	if err != nil {
		return shim.Error(), err
	}
	bidderId, price, err := e.getHighestPriceFromCRDTCOLD(auctionCRDT)
	if err != nil {
		return shim.Error(), err
	}
	logger.InfoLogger.Println("Highest Price", bidderId, price)
	winner := map[string]int32{bidderId: price}
	winnerBuffed, err := json.Marshal(winner)
	if err != nil {
		return shim.Error(), err
	}
	return shim.SuccessWithOutput(winnerBuffed), nil
}

func (e *EnglishAuctionContract) getHighestPriceFromCRDTWarm(auctionCRDT *crdtmanagerv2.CRDTObject) (string, int32, error) {
	_ = auctionCRDT
	highestBidderId := "L"
	var highestBidderPrice int32 = 0
	auctionCRDT.Lock.RLock()
	for bidderId, auctionResult := range auctionCRDT.CrdtObj.Head.Map {
		counter := auctionResult.GCounter
		if counter != nil {
			if counter.CounterValue > highestBidderPrice {
				highestBidderPrice = counter.CounterValue
				highestBidderId = bidderId
			}
		}
	}
	auctionCRDT.Lock.RUnlock()
	return highestBidderId, highestBidderPrice, nil
}

func (e *EnglishAuctionContract) getHighestPriceFromCRDTCOLD(auctionCRDT *protos.CRDTObject) (string, int32, error) {
	_ = auctionCRDT
	highestBidderId := "L"
	var highestBidderPrice int32 = 0
	for bidderId, auctionResult := range auctionCRDT.Head.Map {
		counter := auctionResult.GCounter
		if counter != nil {
			if counter.CounterValue > highestBidderPrice {
				highestBidderPrice = counter.CounterValue
				highestBidderId = bidderId
			}
		}
	}
	return highestBidderId, highestBidderPrice, nil
}
