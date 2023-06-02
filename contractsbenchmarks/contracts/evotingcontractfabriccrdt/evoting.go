package evotingcontractfabriccrdt

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv1"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sort"
	"strconv"
	"strings"
)

func NewPartyVotes(partyId string) *crdtmanagerv1.PartyVotes {
	return &crdtmanagerv1.PartyVotes{
		PartyId: partyId,
		Ballots: map[string]*crdtmanagerv1.Ballot{},
	}
}

type Party struct {
	PartyId   string `json:"party_id"`
	VoteCount int    `json:"vote_count"`
}

func NewParty(partyId string) *Party {
	return &Party{
		PartyId:   partyId,
		VoteCount: 0,
	}
}

func (p *Party) AddVote() {
	p.VoteCount++
}

type Election struct {
	ElectionID string            `json:"election_id"`
	Parties    map[string]*Party `json:"parties"`
}

func NewElection(electionId string, parties map[string]*Party) *Election {
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

func (t *EVotingContract) initElection(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	for _, partyId := range AllParties {
		newParty := NewPartyVotes(electionId + partyId)
		bufferedParty, err := json.Marshal(newParty)
		if err != nil {
			return shim.Error(), err
		}
		shim.PutKeyValueFabricCRDT(electionId+partyId, bufferedParty)
	}

	return shim.Success(), nil
}

func (t *EVotingContract) vote(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	votedPartyId := args[1]
	voterId := args[2]

	_, partyVotes, err := shim.GetJSONCRDTV1Warm(electionId + votedPartyId)
	if err != nil {
		return shim.Error(), err
	}

	partyVotes.Lock.Lock()
	partyVotes.Party.CurrentVoterId = voterId
	partyVotes.Party.CurrentVote = true
	partyVotesBuffered, err := json.Marshal(partyVotes.Party)
	partyVotes.Lock.Unlock()

	if err != nil {
		return shim.Error(), err
	}
	shim.PutKeyValueFabricCRDT(electionId+votedPartyId, partyVotesBuffered)

	for _, partyId := range AllParties {
		if partyId != votedPartyId {
			_, partyVotes, err = shim.GetJSONCRDTV1Warm(electionId + partyId)

			partyVotes.Lock.Lock()
			partyVotes.Party.CurrentVoterId = voterId
			partyVotes.Party.CurrentVote = false
			partyVotesBuffered, err = json.Marshal(partyVotes.Party)
			partyVotes.Lock.Unlock()

			if err != nil {
				return shim.Error(), err
			}
			shim.PutKeyValueFabricCRDT(electionId+partyId, partyVotesBuffered)
		}
	}

	return shim.Success(), nil
}

func (t *EVotingContract) queryParties(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	partyId := args[1]

	_, partyVotes, errParty := shim.GetJSONCRDTV1Warm(electionId + partyId)
	if errParty != nil {
		return shim.Error(), errParty
	}
	voteCount := 0
	partyVotes.Lock.RLock()
	for _, votes := range partyVotes.Party.Ballots {
		lastBallotString := ""
		lastBallotCounter := 0
		for ballotId := range votes.ClientBallots {
			ballotIds := strings.Split(ballotId, ".")
			counter, errParse := strconv.Atoi(ballotIds[0])
			if errParse != nil {

				return shim.Error(), errParse
			}
			if counter > lastBallotCounter {
				lastBallotCounter = counter
				lastBallotString = ballotId
			}
		}
		lastVote := votes.ClientBallots[lastBallotString]
		if lastVote {
			voteCount++
		}
	}
	partyVotes.Lock.RUnlock()

	//logger.InfoLogger.Println("Vote count", voteCount, partyId)
	return shim.Success(), nil
}

func (t *EVotingContract) getElectionResults(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Get election results")

	electionId := args[0]

	allParties := map[string]*Party{}
	for _, partyId := range AllParties {
		newParty := NewParty(electionId + partyId)
		allParties[electionId+partyId] = newParty
	}

	election := NewElection(electionId, allParties)

	for _, partyId := range AllParties {
		_, partyVotes, errParty := shim.GetJSONCRDTV1Warm(electionId + partyId)
		if errParty != nil {
			return shim.Error(), errParty
		}
		partyVotes.Lock.RLock()
		for _, votes := range partyVotes.Party.Ballots {
			lastBallotString := ""
			lastBallotCounter := 0
			for ballotId := range votes.ClientBallots {
				ballotIds := strings.Split(ballotId, ".")
				counter, errParse := strconv.Atoi(ballotIds[0])
				if errParse != nil {
					return shim.Error(), errParse
				}
				if counter > lastBallotCounter {
					lastBallotCounter = counter
					lastBallotString = ballotId
				}
			}
			lastVote := votes.ClientBallots[lastBallotString]
			if lastVote {
				election.Parties[electionId+partyId].AddVote()
			}
		}
		partyVotes.Lock.RUnlock()

	}
	currentResults := election.GetElectionResultsDescending()
	for _, result := range currentResults {
		logger.InfoLogger.Println(result.VoteCount)
	}
	bufferedElection, err := json.Marshal(election)
	if err != nil {
		return shim.Error(), err
	}

	return shim.SuccessWithOutput(bufferedElection), nil
}
