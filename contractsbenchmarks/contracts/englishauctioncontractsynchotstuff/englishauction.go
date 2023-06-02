package englishauctioncontractsynchotstuff

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
	"strings"
)

type Bid struct {
	AuctionID  string `json:"auction_id"`
	BidderId   string `json:"bidder_id"`
	OfferPrice int    `json:"offer_price"`
}

func NewBid(auctionId, bidderId string, offerPrice int) *Bid {
	return &Bid{
		AuctionID:  auctionId,
		BidderId:   bidderId,
		OfferPrice: offerPrice,
	}
}

type Auction struct {
	AuctionID              string `json:"auction_id"`
	CurrentHighestPrice    int    `json:"current_highest_price"`
	CurrentHighestBidderId string `json:"current_highest_bidder_id"`
}

func NewAuction(auctionId string) *Auction {
	return &Auction{
		AuctionID: auctionId,
	}
}

type EnglishAuctionContract struct {
}

func NewContract() *EnglishAuctionContract {
	return &EnglishAuctionContract{}
}

func (e *EnglishAuctionContract) makeCompositeBid(auctionId, bidId string) string {
	return auctionId + "bid" + "-" + bidId
}

func (e *EnglishAuctionContract) makeCompositePartialBid(auctionId string) string {
	return auctionId + "bid"
}

func (e *EnglishAuctionContract) Invoke(shim contractinterface.ShimInterface, proposal *protos.ProposalRequest) (*protos.ProposalResponse, error) {
	proposal.MethodName = strings.ToLower(proposal.MethodName)
	if proposal.MethodName == "createauction" {
		return e.createAuction(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "offerbid" {
		return e.offerBid(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getcurrenthighestprice" {
		return e.getCurrentHighestPrice(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getauctionwinner" {
		return e.getAuctionWinner(shim, proposal.MethodParams)
	}

	return shim.Error(), errors.New("method name not found")
}

func (e *EnglishAuctionContract) createAuction(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	newAuction := NewAuction(auctionId)
	newAuctionBuffered, err := json.Marshal(newAuction)
	if err != nil {
		return shim.Error(), err
	}
	shim.PutKeyValue(auctionId, newAuctionBuffered)

	return shim.Success(), nil
}

func (e *EnglishAuctionContract) offerBid(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	bidderId := args[1]
	offeredPriceString := args[2]

	offeredPrice, err := strconv.Atoi(offeredPriceString)
	if err != nil {
		return shim.Error(), err
	}
	bid := &Bid{}
	bufferedBid, err := shim.GetKeyValueWithVersion(e.makeCompositeBid(auctionId, bidderId))
	if err != nil {
		bid = NewBid(auctionId, bidderId, 0)
	} else {
		err = json.Unmarshal(bufferedBid, bid)
		if err != nil {
			return shim.Error(), err
		}
	}
	bid.OfferPrice += offeredPrice
	bufferedBid, err = json.Marshal(bid)
	if err != nil {
		return shim.Error(), err
	}
	shim.PutKeyValue(e.makeCompositeBid(auctionId, bidderId), bufferedBid)

	bufferedAuction, err := shim.GetKeyValueWithVersion(auctionId)
	if err != nil {
		return shim.Error(), err
	}

	auction := &Auction{}
	err = json.Unmarshal(bufferedAuction, auction)
	if err != nil {
		return shim.Error(), err
	}

	if bid.OfferPrice > auction.CurrentHighestPrice {
		auction.CurrentHighestPrice = bid.OfferPrice
		auction.CurrentHighestBidderId = bid.BidderId
		bufferedAuction, err = json.Marshal(auction)
		if err != nil {
			return shim.Error(), err
		}
		shim.PutKeyValue(auctionId, bufferedAuction)
	}

	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getCurrentHighestPrice(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]

	bufferedAuction, err := shim.GetKeyValueWithVersion(auctionId)
	if err != nil {
		return shim.Error(), err
	}

	auction := &Auction{}
	err = json.Unmarshal(bufferedAuction, auction)
	if err != nil {
		return shim.Error(), err
	}

	//logger.InfoLogger.Println("The current highest bid is:", auction.CurrentHighestBidderId, auction.CurrentHighestPrice)
	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getAuctionWinner(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Getting the auction winner")

	auctionId := args[0]
	bufferedAuction, err := shim.GetKeyValueWithVersion(auctionId)
	if err != nil {
		return shim.Error(), err
	}

	auction := &Auction{}
	err = json.Unmarshal(bufferedAuction, auction)
	if err != nil {
		return shim.Error(), err
	}

	logger.InfoLogger.Println("The winner bid is:", auction.CurrentHighestBidderId, auction.CurrentHighestPrice)

	return shim.SuccessWithOutput(bufferedAuction), nil
}
