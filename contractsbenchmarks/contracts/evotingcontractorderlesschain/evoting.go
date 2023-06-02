package evotingcontractorderlesschain

import (
	"encoding/json"
	"errors"
	"gitlab.lrz.de/orderless/orderlesschain/internal/contract/contractinterface"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sort"
	"strings"
)

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
	if proposal.MethodName == "votewarm" {
		return t.voteWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "votecold" {
		return t.voteCold(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getelectionresultswarm" {
		return t.getElectionResultsWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "getelectionresultscold" {
		return t.getElectionResultsCold(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "querypartieswarm" {
		return t.queryPartiesWarm(shim, proposal.MethodParams)
	}
	if proposal.MethodName == "querypartiescold" {
		return t.queryPartiesCold(shim, proposal.MethodParams)
	}
	return shim.Error(), errors.New("method name not found")
}

func (t *EVotingContract) voteWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	votedPartyId := args[1]
	voterId := args[2]
	voterClockString := args[3]

	operationsList := &protos.CRDTOperationsList{
		CrdtObjectId: electionId + votedPartyId,
		Operations:   []*protos.CRDTOperation{},
	}
	operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
		OperationId:   voterClockString,
		ValueType:     protos.CRDTOperation_MV_REGISTER,
		Value:         "true",
		OperationPath: []string{voterId},
		Clock:         voterClockString,
	})
	shim.PutCRDTOperationsV2Warm(electionId+votedPartyId, operationsList)

	for _, party := range AllParties {
		if party != votedPartyId {
			operationsList = &protos.CRDTOperationsList{
				CrdtObjectId: electionId + party,
				Operations:   []*protos.CRDTOperation{},
			}
			operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
				OperationId:   voterClockString,
				ValueType:     protos.CRDTOperation_MV_REGISTER,
				Value:         "false",
				OperationPath: []string{voterId},
				Clock:         voterClockString,
			})
			shim.PutCRDTOperationsV2Warm(electionId+party, operationsList)
		}
	}

	return shim.Success(), nil
}

func (t *EVotingContract) voteCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	votedPartyId := args[1]
	voterId := args[2]
	voterClockString := args[3]

	operationsList := &protos.CRDTOperationsList{
		CrdtObjectId: electionId + votedPartyId,
		Operations:   []*protos.CRDTOperation{},
	}
	operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
		OperationId:   voterClockString,
		ValueType:     protos.CRDTOperation_MV_REGISTER,
		Value:         "true",
		OperationPath: []string{voterId},
		Clock:         voterClockString,
	})
	shim.PutCRDTOperationsV2Cold(electionId+votedPartyId, operationsList)

	for _, party := range AllParties {
		if party != votedPartyId {
			operationsList = &protos.CRDTOperationsList{
				CrdtObjectId: electionId + party,
				Operations:   []*protos.CRDTOperation{},
			}
			operationsList.Operations = append(operationsList.Operations, &protos.CRDTOperation{
				OperationId:   voterClockString,
				ValueType:     protos.CRDTOperation_MV_REGISTER,
				Value:         "false",
				OperationPath: []string{voterId},
				Clock:         voterClockString,
			})
			shim.PutCRDTOperationsV2Cold(electionId+party, operationsList)
		}
	}

	return shim.Success(), nil
}

func (t *EVotingContract) queryPartiesWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	partyId := args[1]

	partyCRDT, errCRDT := shim.GetCRDTObjectV2Warm(electionId + partyId)
	if errCRDT != nil {
		return shim.Error(), errCRDT
	}
	partyVote := 0
	partyCRDT.Lock.RLock()
	for _, voterResult := range partyCRDT.CrdtObj.Head.Map {
		if voterResult.MvRegister != nil &&
			len(voterResult.MvRegister.RegisterValue) == 1 && voterResult.MvRegister.RegisterValue[0] {
			partyVote++
		}
	}
	partyCRDT.Lock.RUnlock()
	//logger.InfoLogger.Println(partyVote)
	return shim.Success(), nil
}

func (t *EVotingContract) queryPartiesCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	electionId := args[0]
	partyId := args[1]

	partyCRDT, errCRDT := shim.GetCRDTObjectV2Cold(electionId + partyId)
	if errCRDT != nil {
		return shim.Error(), errCRDT
	}
	partyVote := 0
	for _, voterResult := range partyCRDT.Head.Map {
		if voterResult.MvRegister != nil &&
			len(voterResult.MvRegister.RegisterValue) == 1 && voterResult.MvRegister.RegisterValue[0] {
			partyVote++
		}
	}
	//logger.InfoLogger.Println(partyVote)
	return shim.Success(), nil
}

func (t *EVotingContract) getElectionResultsWarm(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Get election results Warm")

	electionId := args[0]

	var allParties []*Party
	for _, partyId := range AllParties {
		newParty := NewParty(partyId)
		allParties = append(allParties, newParty)
	}

	tempElection := NewElection(electionId, allParties)

	for _, party := range tempElection.Parties {
		partyCRDT, errCRDT := shim.GetCRDTObjectV2Warm(electionId + party.PartyId)
		if errCRDT != nil {
			return shim.Error(), errCRDT
		}
		partyCRDT.Lock.RLock()
		if partyCRDT.CrdtObj.Head != nil {
			for _, voterResult := range partyCRDT.CrdtObj.Head.Map {
				if voterResult.MvRegister != nil &&
					len(voterResult.MvRegister.RegisterValue) == 1 && voterResult.MvRegister.RegisterValue[0] {
					party.AddVote()
				}
			}
		}
		partyCRDT.Lock.RUnlock()
	}

	currentResults := tempElection.GetElectionResultsDescending()
	for _, result := range currentResults {
		logger.InfoLogger.Println(result.VoteCount)
	}
	electionToSend := NewElection(electionId, currentResults)
	bufferedElection, err := json.Marshal(electionToSend)
	if err != nil {
		return shim.Error(), err
	}
	return shim.SuccessWithOutput(bufferedElection), nil
}

func (t *EVotingContract) getElectionResultsCold(shim contractinterface.ShimInterface, args []string) (*protos.ProposalResponse, error) {
	logger.InfoLogger.Println("Get election results Cold")

	electionId := args[0]

	var allParties []*Party
	for _, partyId := range AllParties {
		newParty := NewParty(partyId)
		allParties = append(allParties, newParty)
	}

	tempElection := NewElection(electionId, allParties)

	for _, party := range tempElection.Parties {
		partyCRDT, errCRDT := shim.GetCRDTObjectV2Cold(electionId + party.PartyId)
		if errCRDT != nil {
			return shim.Error(), errCRDT
		}
		if partyCRDT.Head != nil {
			for _, voterResult := range partyCRDT.Head.Map {
				if voterResult.MvRegister != nil &&
					len(voterResult.MvRegister.RegisterValue) == 1 && voterResult.MvRegister.RegisterValue[0] {
					party.AddVote()
				}
			}
		}
	}
	currentResults := tempElection.GetElectionResultsDescending()
	for _, result := range currentResults {
		logger.InfoLogger.Println(result.VoteCount)
	}
	electionToSend := NewElection(electionId, currentResults)
	bufferedElection, err := json.Marshal(electionToSend)
	if err != nil {
		return shim.Error(), err
	}
	return shim.SuccessWithOutput(bufferedElection), nil
}
