package crdtmanagerv1

import (
	"errors"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strings"
	"sync"
)

type JSONDocs struct {
	JSONDocs        map[string]*protos.Doc
	JSONDocsAuction map[string]*AuctionContainer
	JSONDocsParty   map[string]*PartyContainer
	lock            *sync.Mutex
}

type AuctionContainer struct {
	Auction *Auction
	Lock    *sync.RWMutex
}

type PartyContainer struct {
	Party *PartyVotes
	Lock  *sync.RWMutex
}

type CRDTCollections struct {
	JSONDocs *JSONDocs
}

func NewCRDTCollections() *CRDTCollections {
	return &CRDTCollections{
		JSONDocs: &JSONDocs{
			JSONDocs:        map[string]*protos.Doc{},
			JSONDocsAuction: map[string]*AuctionContainer{},
			JSONDocsParty:   map[string]*PartyContainer{},
			lock:            &sync.Mutex{},
		},
	}
}

func (j *JSONDocs) GetJSONDocs(key string) (*protos.Doc, bool) {
	i, ok := j.JSONDocs[key]
	return i, ok
}

func NewNode(opId string) *protos.Node {
	return &protos.Node{
		Deps: []string{},
		OpID: opId,
		Hmap: make(map[string]*protos.Node),
		List: []*protos.Node{},
		Reg:  make(map[string]string),
	}
}

func Init(id string) *protos.Doc {
	headNode := NewNode("")
	return &protos.Doc{
		Id:              id,
		Clocks:          []*protos.Clock{},
		Clock:           NewClock(id),
		OperationsId:    []string{},
		Head:            headNode,
		OperationBuffer: []*protos.Operation{},
	}
}

type Manager struct {
	CRDTObjs *CRDTCollections
}

func NewManager() *Manager {
	return &Manager{
		CRDTObjs: NewCRDTCollections(),
	}
}

func (m *Manager) SetJSONCRDTValueWarm(contractName, key string, value interface{}) {
	m.CRDTObjs.JSONDocs.lock.Lock()
	defer m.CRDTObjs.JSONDocs.lock.Unlock()

	switch value.(type) {
	case *Auction:
		var auctionContainer *AuctionContainer
		var ok bool
		if auctionContainer, ok = m.CRDTObjs.JSONDocs.JSONDocsAuction[contractName+key]; !ok {
			auctionContainer = &AuctionContainer{
				Auction: value.(*Auction),
				Lock:    &sync.RWMutex{},
			}
			m.CRDTObjs.JSONDocs.JSONDocsAuction[contractName+key] = auctionContainer
		}
		auctionContainer.Lock.Lock()
		auctionContainer.Auction = value.(*Auction)
		auctionContainer.Lock.Unlock()
	case *PartyVotes:
		var partyContainer *PartyContainer
		var ok bool
		if partyContainer, ok = m.CRDTObjs.JSONDocs.JSONDocsParty[contractName+key]; !ok {
			partyContainer = &PartyContainer{
				Party: value.(*PartyVotes),
				Lock:  &sync.RWMutex{},
			}
			m.CRDTObjs.JSONDocs.JSONDocsParty[contractName+key] = partyContainer
		}
		partyContainer.Lock.Lock()
		partyContainer.Party = value.(*PartyVotes)
		partyContainer.Lock.Unlock()
	}
}

func (m *Manager) GetJSONCRDTValueWarm(contractName, key string) (*AuctionContainer, *PartyContainer, error) {
	m.CRDTObjs.JSONDocs.lock.Lock()
	defer m.CRDTObjs.JSONDocs.lock.Unlock()

	if strings.Contains(contractName, "auction") {
		var auctionContainer *AuctionContainer
		var ok bool
		if auctionContainer, ok = m.CRDTObjs.JSONDocs.JSONDocsAuction[contractName+key]; !ok {
			return nil, nil, errors.New("object was not found")
		}
		return auctionContainer, nil, nil
	} else if strings.Contains(contractName, "voting") {
		var partyContainer *PartyContainer
		var ok bool
		if partyContainer, ok = m.CRDTObjs.JSONDocs.JSONDocsParty[contractName+key]; !ok {
			return nil, nil, errors.New("object was not found")
		}
		return nil, partyContainer, nil
	}
	return nil, nil, errors.New("object was not found")
}
