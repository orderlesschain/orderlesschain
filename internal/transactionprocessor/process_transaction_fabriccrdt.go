package transactionprocessor

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"gitlab.lrz.de/orderless/orderlesschain/internal/crdtmanagerv1"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transaction"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"sort"
	"sync"
)

func (p *Processor) signProposalResponseFabricCRDT(proposalResponse *protos.ProposalResponse) (*protos.ProposalResponse, error) {
	sort.Slice(proposalResponse.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return proposalResponse.ReadWriteSet.ReadKeys.ReadKeys[i].Key < proposalResponse.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	marshalledReadKeys, err := proto.Marshal(proposalResponse.ReadWriteSet.ReadKeys)
	if err != nil {
		return p.makeFailedProposal(proposalResponse.ProposalId), nil
	}
	proposalResponse.NodeSignature = p.signer.Sign(marshalledReadKeys)
	return proposalResponse, nil
}

func (p *Processor) preProcessValidateReadSetFabricCRDT(tx *protos.Transaction, wg *sync.WaitGroup) {
	sort.Slice(tx.ReadWriteSet.ReadKeys.ReadKeys, func(i, j int) bool {
		return tx.ReadWriteSet.ReadKeys.ReadKeys[i].Key < tx.ReadWriteSet.ReadKeys.ReadKeys[j].Key
	})
	txDigest, err := proto.Marshal(tx.ReadWriteSet.ReadKeys)
	if err != nil {
		tx.Status = protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION
		wg.Done()
		return
	}
	for node, nodeSign := range tx.NodeSignatures {
		err = p.signer.Verify(node, txDigest, nodeSign)
		if err != nil {
			tx.Status = protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION
			wg.Done()
			return
		}
	}
	wg.Done()
}

func (p *Processor) ProcessProposalFabricCRDTStream(proposal *protos.ProposalRequest) {
	p.transactionProfiler.AddEndorseStart(&proposal.ProposalId)
	p.txJournal.AddProposalToQueue(proposal)
}

func (p *Processor) runProposalQueueProcessingFabricCRDT() {
	for {
		proposals := <-p.txJournal.DequeuedProposalsChan
		go p.processDequeuedProposalsFabricCRDT(proposals)
	}
}

func (p *Processor) processDequeuedProposalsFabricCRDT(proposals *transaction.DequeuedProposals) {
	for _, proposal := range proposals.DequeuedProposals {
		go p.processProposalFabricCRDT(proposal)
	}
}

func (p *Processor) processProposalFabricCRDT(proposal *protos.ProposalRequest) {
	response, err := p.executeContract(proposal)
	if err != nil {
		p.sendProposalResponseToSubscriber(proposal.ClientId, response)
		return
	}
	response, err = p.signProposalResponseFabricCRDT(response)
	p.sendProposalResponseToSubscriber(proposal.ClientId, response)
}

func (p *Processor) processBlockFabricCRDT(block *protos.Block) {
	wg := &sync.WaitGroup{}
	wg.Add(len(block.Transactions))
	for _, tx := range block.Transactions {
		go p.preProcessValidateReadSetFabricCRDT(tx, wg)
	}
	wg.Wait()
	for _, tx := range block.Transactions {
		if tx.Status == protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
			go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION, block.ThisBlockHash))
		}
	}
	p.updateCRDTObjectFabricCRDT(block)
	convertedValues := map[string][]byte{}
	for _, tx := range block.Transactions {
		if tx.Status != protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
			err := p.validateMVCCFabricCRDT(tx)
			if err != nil {
				go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_MVCC, block.ThisBlockHash))
			}
			isCRDT := p.concludeCRDTObjectsFabricCRDT(tx, convertedValues)
			response, err := p.commitTransactionFabricCRDT(tx, isCRDT)
			if err == nil {
				response.BlockHeader = block.ThisBlockHash
				go p.sendTransactionResponseToSubscriber(tx.ClientId, response)
			} else {
				go p.sendTransactionResponseToSubscriber(tx.ClientId, p.makeFailedTransactionResponse(tx.TransactionId, protos.TransactionStatus_FAILED_DATABASE, block.ThisBlockHash))
			}
		}
	}
}

func (p *Processor) validateMVCCFabricCRDT(tx *protos.Transaction) error {
	// This is done like this, because FabricCRDT supports execution of both types of CRDT and nonCRDT transactions
	if len(tx.ReadWriteSet.WriteKeyValues.WriteKeyValues) == 0 {
		return nil
	}
	for _, writeKey := range tx.ReadWriteSet.WriteKeyValues.WriteKeyValues {
		if writeKey.WriteType == protos.WriteKeyValue_FABRICCRDT {
			return nil
		}
	}
	dbOp := p.sharedShimResources.DBConnections[tx.ContractName]
	for _, readKey := range tx.ReadWriteSet.ReadKeys.ReadKeys {
		if readKey.VersionNumber > 0 && dbOp.GetKeyCurrentVersion(readKey.Key) != readKey.VersionNumber {
			return errors.New("MVCC validation error")
		}
	}
	return nil
}

func (p *Processor) updateCRDTObjectFabricCRDT(block *protos.Block) {
	for _, tx := range block.Transactions {
		if tx.Status != protos.TransactionStatus_FAILED_SIGNATURE_VALIDATION {
			for _, nsRWSet := range tx.ReadWriteSet.WriteKeyValues.WriteKeyValues {
				if nsRWSet.WriteType == protos.WriteKeyValue_FABRICCRDT {
					crdtDoc, ok := p.crdtManagerFabricCRDT.CRDTObjs.JSONDocs.JSONDocs[nsRWSet.Key]
					if !ok {
						crdtDoc = crdtmanagerv1.Init(nsRWSet.Key)
						dbOp := p.sharedShimResources.DBConnections[tx.ContractName]
						if bufferedJson, err := dbOp.GetKeyValueNoVersion(nsRWSet.Key); err == nil {
							if err = crdtmanagerv1.MakeCRDTBasedOnType(tx.ContractName, bufferedJson, crdtDoc); err != nil {
								crdtDoc = crdtmanagerv1.Init(nsRWSet.Key)
							}
						}
						p.crdtManagerFabricCRDT.CRDTObjs.JSONDocs.JSONDocs[nsRWSet.Key] = crdtDoc
					}
					err := crdtmanagerv1.MakeCRDTBasedOnType(tx.ContractName, nsRWSet.Value, crdtDoc)
					if err != nil {
						logger.ErrorLogger.Println(err)
					}
				}
			}
		}
	}
}

func (p *Processor) concludeCRDTObjectsFabricCRDT(tx *protos.Transaction, convertedValues map[string][]byte) bool {
	isCRDT := false
	for _, nsRWSet := range tx.ReadWriteSet.WriteKeyValues.WriteKeyValues {
		if nsRWSet.WriteType == protos.WriteKeyValue_FABRICCRDT {
			isCRDT = true
			if _, ok := p.crdtManagerFabricCRDT.CRDTObjs.JSONDocs.GetJSONDocs(nsRWSet.Key); ok {
				if convertedValue, okExist := convertedValues[nsRWSet.Key]; okExist {
					nsRWSet.Value = convertedValue
				} else {
					if convertedValueNew, err := crdtmanagerv1.DocToJsonByteBasedOnType(tx.ContractName, nsRWSet.Key,
						p.crdtManagerFabricCRDT.CRDTObjs.JSONDocs.JSONDocs[nsRWSet.Key], p.crdtManagerFabricCRDT); err != nil {
						logger.ErrorLogger.Println(err)
					} else {
						convertedValues[nsRWSet.Key] = convertedValueNew
						nsRWSet.Value = convertedValueNew
					}
				}
			}
		}
	}
	return isCRDT
}

func (p *Processor) commitTransactionFabricCRDT(tx *protos.Transaction, isCRDT bool) (*protos.TransactionResponse, error) {
	if err := p.addTransactionToDatabaseFabricCRDT(tx, isCRDT); err != nil {
		return nil, err
	}
	return p.makeSuccessTransactionResponse(tx.TransactionId, []byte{}), nil
}

func (p *Processor) addTransactionToDatabaseFabricCRDT(tx *protos.Transaction, isCRDT bool) error {
	dbOp := p.sharedShimResources.DBConnections[tx.ContractName]
	if isCRDT {
		for _, keyValue := range tx.ReadWriteSet.WriteKeyValues.WriteKeyValues {
			if err := dbOp.PutKeyValueNoVersion(keyValue.Key, keyValue.Value); err != nil {
				return err
			}
		}
	} else {
		for _, keyValue := range tx.ReadWriteSet.WriteKeyValues.WriteKeyValues {
			if err := dbOp.PutKeyValueWithVersion(keyValue.Key, keyValue.Value); err != nil {
				return err
			}
		}
	}
	return nil
}
