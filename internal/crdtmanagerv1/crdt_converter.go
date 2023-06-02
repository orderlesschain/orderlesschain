package crdtmanagerv1

import (
	"encoding/json"
	"errors"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
)

func DocToJsonByteBasedOnType(contractName, key string, convertedDoc *protos.Doc, manager *Manager) ([]byte, error) {
	if contractName == "evotingcontractfabriccrdt" {
		return docToJsonBytePartyVote(contractName, key, convertedDoc, manager)
	}
	if contractName == "englishauctioncontractfabriccrdt" {
		return docToJsonByteAuction(contractName, key, convertedDoc, manager)
	}
	return nil, errors.New("json converter not found")
}

func docToJsonBytePartyVote(contractName, key string, convertedDoc *protos.Doc, manager *Manager) ([]byte, error) {
	partyVote := &PartyVotes{
		Ballots: map[string]*Ballot{},
	}
	if convertedDoc.Head.Reg != nil {
		if partyId, ok := convertedDoc.Head.Reg["party_id"]; ok {
			partyVote.PartyId = partyId
		}
	}
	if convertedDoc.Head.Hmap != nil {
		for voterId, votes := range convertedDoc.Head.Hmap {
			tempVote := &Ballot{
				ClientBallots: map[string]bool{},
			}
			for voterClock, vote := range votes.Hmap {
				voteBool, _ := strconv.ParseBool(vote.Reg["vote"])
				tempVote.ClientBallots[voterClock] = voteBool
			}
			partyVote.Ballots[voterId] = tempVote
		}
	}
	manager.SetJSONCRDTValueWarm(contractName, key, partyVote)
	return json.Marshal(partyVote)
}

func docToJsonByteAuction(contractName, key string, convertedDoc *protos.Doc, manager *Manager) ([]byte, error) {
	auction := &Auction{
		Bids: map[string]*Bid{},
	}
	if convertedDoc.Head.Reg != nil {
		if auctionId, ok := convertedDoc.Head.Reg["auction_id"]; ok {
			auction.AuctionID = auctionId
		}
	}
	if convertedDoc.Head.Hmap != nil {
		for bidderId, bids := range convertedDoc.Head.Hmap {
			tempBids := &Bid{
				ClientBids: map[string]int{},
			}
			for bidderClock, bid := range bids.Hmap {
				bidInt, _ := strconv.Atoi(bid.Reg["bid"])
				tempBids.ClientBids[bidderClock] = bidInt
			}
			auction.Bids[bidderId] = tempBids
		}
	}
	manager.SetJSONCRDTValueWarm(contractName, key, auction)
	return json.Marshal(auction)
}
