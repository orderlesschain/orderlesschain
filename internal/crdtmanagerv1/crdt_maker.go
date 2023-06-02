package crdtmanagerv1

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strconv"
)

type Bid struct {
	ClientBids map[string]int `json:"client_bids"`
}

func MakeCRDTBasedOnType(contractName string, writeValue []byte, convertedDoc *protos.Doc) error {
	if contractName == "evotingcontractfabriccrdt" {
		return makePartyVote(writeValue, convertedDoc)
	}
	if contractName == "englishauctioncontractfabriccrdt" {
		return makeAuction(writeValue, convertedDoc)
	}
	return errors.New("crdt maker not found")
}

type PartyVotes struct {
	PartyId        string             `json:"party_id"`
	CurrentVoterId string             `json:"current_voter_id"`
	CurrentVote    bool               `json:"current_vote"`
	Ballots        map[string]*Ballot `json:"ballots"`
}

type Ballot struct {
	ClientBallots map[string]bool `json:"client_ballots"`
}

func makePartyVote(writeValue []byte, convertedDoc *protos.Doc) error {
	partyVote := &PartyVotes{}
	if err := json.Unmarshal(writeValue, partyVote); err != nil {
		return err
	}
	if convertedDoc.Head.Reg == nil {
		convertedDoc.Head.Reg = map[string]string{}
	}
	convertedDoc.Head.Reg["party_id"] = partyVote.PartyId
	if convertedDoc.Head.Hmap == nil {
		convertedDoc.Head.Hmap = map[string]*protos.Node{}
	}
	for voterId, votes := range partyVote.Ballots {
		voterVotes, ok := convertedDoc.Head.Hmap[voterId]
		if !ok {
			voterVotes = &protos.Node{
				Hmap: map[string]*protos.Node{},
			}
			convertedDoc.Head.Hmap[voterId] = voterVotes
		}
		for voterClock, vote := range votes.ClientBallots {
			if _, ok = voterVotes.Hmap[voterClock]; !ok {
				voterVotes.Hmap[voterClock] = &protos.Node{
					Reg: map[string]string{"vote": strconv.FormatBool(vote)},
				}
			}
			if err := UpdateClock(convertedDoc.Clock, voterClock); err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
	}

	if len(partyVote.CurrentVoterId) > 0 {
		Tick(convertedDoc.Clock)
		opDep := ConvertClockToString(convertedDoc.Clock)
		convertedDoc.OperationsId = append(convertedDoc.OperationsId, opDep)
		convertedDoc.Head.Deps = append(convertedDoc.Head.Deps, opDep)

		voterVotes, ok := convertedDoc.Head.Hmap[partyVote.CurrentVoterId]
		if !ok {
			voterVotes = &protos.Node{
				Hmap: map[string]*protos.Node{},
			}
			convertedDoc.Head.Hmap[partyVote.CurrentVoterId] = voterVotes
		}
		voterVotes.Hmap[ConvertClockToString(convertedDoc.Clock)] = &protos.Node{
			Reg: map[string]string{"vote": strconv.FormatBool(partyVote.CurrentVote)},
		}
	}

	return nil
}

type Auction struct {
	AuctionID       string          `json:"auction_id"`
	CurrentBidderId string          `json:"current_bidder_id"`
	CurrentBid      int             `json:"current_bid"`
	Bids            map[string]*Bid `json:"bids"`
}

func makeAuction(writeValue []byte, convertedDoc *protos.Doc) error {
	auction := &Auction{}
	if err := json.Unmarshal(writeValue, auction); err != nil {
		return err
	}
	if convertedDoc.Head.Reg == nil {
		convertedDoc.Head.Reg = map[string]string{}
	}
	convertedDoc.Head.Reg["auction_id"] = auction.AuctionID
	if convertedDoc.Head.Hmap == nil {
		convertedDoc.Head.Hmap = map[string]*protos.Node{}
	}
	for bidderId, bids := range auction.Bids {
		bidderBids, ok := convertedDoc.Head.Hmap[bidderId]
		if !ok {
			bidderBids = &protos.Node{
				Hmap: map[string]*protos.Node{},
			}
			convertedDoc.Head.Hmap[bidderId] = bidderBids
		}
		for bidderClock, bid := range bids.ClientBids {
			if _, ok = bidderBids.Hmap[bidderClock]; !ok {
				bidderBids.Hmap[bidderClock] = &protos.Node{
					Reg: map[string]string{"bid": strconv.Itoa(bid)},
				}
			}
			if err := UpdateClock(convertedDoc.Clock, bidderClock); err != nil {
				logger.ErrorLogger.Println(err)
			}
		}
	}

	if len(auction.CurrentBidderId) > 0 {
		Tick(convertedDoc.Clock)
		opDep := ConvertClockToString(convertedDoc.Clock)
		convertedDoc.OperationsId = append(convertedDoc.OperationsId, opDep)
		convertedDoc.Head.Deps = append(convertedDoc.Head.Deps, opDep)

		bidderBids, ok := convertedDoc.Head.Hmap[auction.CurrentBidderId]
		if !ok {
			bidderBids = &protos.Node{
				Hmap: map[string]*protos.Node{},
			}
			convertedDoc.Head.Hmap[auction.CurrentBidderId] = bidderBids
		}
		bidderBids.Hmap[ConvertClockToString(convertedDoc.Clock)] = &protos.Node{
			Reg: map[string]string{"bid": strconv.Itoa(auction.CurrentBid)},
		}
	}

	return nil
}
