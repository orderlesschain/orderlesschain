package englishauctioncontractfabriccrdt

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv1"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
	"strings"
)

func NewAuction(auctionId string) *crdtmanagerv1.Auction {
	return &crdtmanagerv1.Auction{
		AuctionID: auctionId,
		Bids:      map[string]*crdtmanagerv1.Bid{},
	}
}

type EnglishAuctionContract struct {
}

func NewContract() *EnglishAuctionContract {
	return &EnglishAuctionContract{}
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
	shim.PutKeyValueFabricCRDT(auctionId, newAuctionBuffered)

	return shim.Success(), nil
}

func (e *EnglishAuctionContract) offerBid(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	bidderId := args[1]
	offeredPriceString := args[2]

	offeredPriceStringInt, err := strconv.Atoi(offeredPriceString)
	if err != nil {
		return shim.Error(), err
	}
	auction, _, err := shim.GetJSONCRDTV1Warm(auctionId)

	auction.Lock.Lock()
	auction.Auction.CurrentBidderId = bidderId
	auction.Auction.CurrentBid = offeredPriceStringInt
	auctionBuffered, err := json.Marshal(auction.Auction)
	auction.Lock.Unlock()

	if err != nil {
		return shim.Error(), err
	}
	shim.PutKeyValueFabricCRDT(auctionId, auctionBuffered)

	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getCurrentHighestPrice(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	auctionId := args[0]
	bidderId, price, err := e.getHighestPriceFromCRDT(shim, auctionId)
	if err != nil {
		return shim.Error(), err
	}
	_, _ = bidderId, price
	//logger.InfoLogger.Println("Highest Price", bidderId, price)
	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getAuctionWinner(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Getting the auction winner")
	auctionId := args[0]
	bidderId, price, err := e.getHighestPriceFromCRDT(shim, auctionId)
	if err != nil {
		return shim.Error(), err
	}
	logger.InfoLogger.Println("Highest Price", bidderId, price)
	return shim.Success(), nil
}

func (e *EnglishAuctionContract) getHighestPriceFromCRDT(shim contractinterface.ShimInterface, auctionId string) (string, int, error) {
	auction, _, err := shim.GetJSONCRDTV1Warm(auctionId)

	highestBidderId := "L"
	highestBidderPrice := 0
	auction.Lock.RLock()
	if auction.Auction != nil {
		for bidderId, bids := range auction.Auction.Bids {
			lastBidString := ""
			lastBidCounter := 0
			for bidId := range bids.ClientBids {
				bidIds := strings.Split(bidId, ".")
				counter, errParse := strconv.Atoi(bidIds[0])
				if errParse != nil {
					return "", 0, err
				}
				if counter > lastBidCounter {
					lastBidCounter = counter
					lastBidString = bidId
				}
			}
			lastBid := bids.ClientBids[lastBidString]
			if lastBid > highestBidderPrice {
				highestBidderPrice = lastBid
				highestBidderId = bidderId
			}
		}
	}
	auction.Lock.RUnlock()
	return highestBidderId, highestBidderPrice, nil
}
