package evotingcontractsynchotstuff

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sort"
	"strings"
)

type Ballot struct {
	VoterId      string `json:"voter_id"`
	ElectionID   string `json:"election_id"`
	PreviousVote string `json:"previous_vote"`
}

func NewBallot(voterId, electionId string) *Ballot {
	return &Ballot{
		VoterId:    voterId,
		ElectionID: electionId,
	}
}

func (b *Ballot) IfPreviouslyVoted() (bool, string) {
	if len(b.PreviousVote) > 0 {
		return true, b.PreviousVote
	}
	return false, ""
}

func (b *Ballot) VoteForParty(partyId string) {
	b.PreviousVote = partyId
}

type Party struct {
	PartyId   string `json:"party_id"`
	VoteCount int    `json:"vote_count"`
}

func NewParty(partyId string) *Party {
	return &Party{
		PartyId: partyId,
	}
}

func (p *Party) AddVote() {
	p.VoteCount++
}

func (p *Party) RemoveVote() {
	p.VoteCount--
}

type Election struct {
	ElectionID string   `json:"election_id"`
	Parties    []*Party `json:"parties"`
}

func NewElection(electionId string, parties []*Party) *Election {
	return &Election{
		ElectionID: electionId,
		Parties:    parties,
	}
}

func (e *Election) GetElectionResultsDescending() []*Party {
	var parties []*Party
	for _, party := range e.Parties {
		parties = append(parties, party)
	}
	sort.Slice(parties, func(i, j int) bool {
		return parties[i].VoteCount < parties[j].VoteCount
	})
	return parties
}

type EVotingContract struct {
}

func NewContract() *EVotingContract {
	return &EVotingContract{}
}

func (t *EVotingContract) Invoke(shim contractinterface.ShimInterface, proposal *protos.ProposalRequest) (*protos.ProposalResponse, error) {
	proposal.MethodName = strings.ToLower(proposal.MethodName)
	if proposal.MethodName == "initelection" {
		return t.initElection(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "vote" {
		return t.vote(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "queryparties" {
		return t.queryParties(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getelectionresults" {
		return t.getElectionResults(shim, proposal.MethodParams)
	}
	return shim.Error(), errors.New("method name not found")
}

func (t *EVotingContract) makeCompositeBallot(electionId, ballotId string) string {
	return electionId + "ballot" + "-" + ballotId
}

func (t *EVotingContract) initElection(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]

	var allParties []*Party
	for _, partyId := range AllParties {
		newParty := NewParty(partyId)
		bufferedParty, err := json.Marshal(newParty)
		if err != nil {
			return shim.Error(), err
		}
		shim.PutKeyValue(electionId+partyId, bufferedParty)
		allParties = append(allParties, newParty)
	}

	newElection := NewElection(electionId, allParties)
	bufferedElection, err := json.Marshal(newElection)
	if err != nil {
		return shim.Error(), err
	}
	shim.PutKeyValue(electionId, bufferedElection)

	return shim.Success(), nil
}

func (t *EVotingContract) vote(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	partyId := args[1]
	voterId := args[2]

	ballot := &Ballot{}
	bufferedBallot, err := shim.GetKeyValueWithVersion(t.makeCompositeBallot(electionId, voterId))
	if err != nil {
		ballot = NewBallot(voterId, electionId)
	} else {
		err = json.Unmarshal(bufferedBallot, ballot)
		if err != nil {
			return shim.Error(), err
		}
	}
	hasVoted, previousVote := ballot.IfPreviouslyVoted()
	if hasVoted {
		bufferedParty, err := shim.GetKeyValueWithVersion(electionId + previousVote)
		if err != nil {
			return shim.Error(), err
		}
		party := &Party{}
		err = json.Unmarshal(bufferedParty, party)
		if err != nil {
			return shim.Error(), err
		}
		party.RemoveVote()
		bufferedParty, err = json.Marshal(party)
		if err != nil {
			return shim.Error(), err
		}
		shim.PutKeyValue(electionId+previousVote, bufferedParty)
	}
	ballot.VoteForParty(partyId)

	bufferedParty, err := shim.GetKeyValueWithVersion(electionId + partyId)
	if err != nil {
		return shim.Error(), err
	}
	party := &Party{}
	err = json.Unmarshal(bufferedParty, party)
	if err != nil {
		return shim.Error(), err
	}
	party.AddVote()

	bufferedBallot, err = json.Marshal(ballot)
	if err != nil {
		return shim.Error(), err
	}
	shim.PutKeyValue(t.makeCompositeBallot(electionId, voterId), bufferedBallot)

	bufferedParty, err = json.Marshal(party)
	if err != nil {
		return shim.Error(), err
	}
	shim.PutKeyValue(electionId+partyId, bufferedParty)

	return shim.Success(), nil
}

func (t *EVotingContract) queryParties(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	partyId := args[1]

	bufferedParty, errCRDT := shim.GetKeyValueWithVersion(electionId + partyId)
	if errCRDT != nil {
		return shim.Success(), nil
	}
	party := &Party{}
	err := json.Unmarshal(bufferedParty, party)
	if err != nil {
		return shim.Success(), nil
	}
	return shim.Success(), nil
}

func (t *EVotingContract) getElectionResults(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Get election results")
	electionId := args[0]

	var parties []*Party
	electionToSend := NewElection(electionId, parties)

	for _, partyId := range AllParties {
		bufferedParty, errCRDT := shim.GetKeyValueWithVersion(electionId + partyId)
		if errCRDT != nil {
			return shim.Error(), errCRDT
		}
		party := &Party{}
		err := json.Unmarshal(bufferedParty, party)
		if err != nil {
			return shim.Error(), err
		}
		electionToSend.Parties = append(electionToSend.Parties, party)
	}
	currentResults := electionToSend.GetElectionResultsDescending()
	electionToSend.Parties = currentResults
	for _, result := range currentResults {
		logger.InfoLogger.Println(result.VoteCount)
	}
	bufferedElection, err := json.Marshal(electionToSend)
	if err != nil {
		return shim.Error(), err
	}
	return shim.SuccessWithOutput(bufferedElection), nil
}
