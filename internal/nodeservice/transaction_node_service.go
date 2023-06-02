package nodeservice

import (
	"context"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/cpumemoryprofiling"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/keygenerator"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/profiling"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"io"
	"os"
)

type TransactionService struct {
	transactionProcessor *transactionprocessor.Processor
	publicPrivateKey     *keygenerator.RSAKey
	cpuMemoryProfiler    *cpumemoryprofiling.CPUMemoryProfiler
}

func NewTransactionService() *TransactionService {
	return &TransactionService{
		transactionProcessor: transactionprocessor.InitTransactionProcessor(),
		publicPrivateKey:     keygenerator.LoadPublicPrivateKeyFromFile(),
		cpuMemoryProfiler:    cpumemoryprofiling.InitCPUMemoryProfiler(),
	}
}

func (t *TransactionService) GetPublicKey(_ context.Context, _ *protos.Empty) (*protos.PublicKeyResponse, error) {
	return &protos.PublicKeyResponse{
		PublicKey: t.publicPrivateKey.PublicKeyString,
		NodeId:    config.Config.UUID,
	}, nil
}

func (t *TransactionService) ChangeModeRestart(_ context.Context, opm *protos.OperationMode) (*protos.Empty, error) {
	go config.UpdateModeAndRestart(opm)
	return &protos.Empty{}, nil
}

func (t *TransactionService) FailureCommand(_ context.Context, fc *protos.FailureCommandMode) (*protos.Empty, error) {
	go t.transactionProcessor.SetFailureCommand(fc)
	return &protos.Empty{}, nil
}

func (t *TransactionService) LoadOtherNodesPublicKeys() {
	t.transactionProcessor.LoadOtherNodesPublicKeys()
}

func (t *TransactionService) ProcessProposalOrderlessChainStream(stream protos.TransactionService_ProcessProposalOrderlessChainStreamServer) error {
	for {
		proposal, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.Empty{})
		}
		if err != nil {
			return err
		}
		if t.transactionProcessor.ShouldFailByzantineNetwork() {
			continue
		}
		t.transactionProcessor.ProcessProposalOrderlessChainStream(proposal)
	}
}

func (t *TransactionService) CommitOrderlessChainTransactionStream(stream protos.TransactionService_CommitOrderlessChainTransactionStreamServer) error {
	for {
		transaction, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.Empty{})
		}
		if err != nil {
			return err
		}
		if t.transactionProcessor.ShouldFailByzantineNetwork() {
			continue
		}
		t.transactionProcessor.ProcessTransactionOrderlessChainStream(transaction)
	}
}

func (t *TransactionService) ProcessProposalFabricStream(stream protos.TransactionService_ProcessProposalFabricStreamServer) error {
	for {
		proposal, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.Empty{})
		}
		if err != nil {
			return err
		}
		t.transactionProcessor.ProcessProposalFabricStream(proposal)
	}
}

func (t *TransactionService) ProcessProposalFabricCRDTStream(stream protos.TransactionService_ProcessProposalFabricCRDTStreamServer) error {
	for {
		proposal, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.Empty{})
		}
		if err != nil {
			return err
		}
		t.transactionProcessor.ProcessProposalFabricCRDTStream(proposal)
	}
}

func (t *TransactionService) SubscribeProposalResponse(subscription *protos.ProposalResponseEventSubscription,
	stream protos.TransactionService_SubscribeProposalResponseServer) error {
	return t.transactionProcessor.ProposalResponseSubscription(subscription, stream)
}

func (t *TransactionService) SubscribeTransactionResponse(subscription *protos.TransactionResponseEventSubscription,
	stream protos.TransactionService_SubscribeTransactionResponseServer) error {
	return t.transactionProcessor.TransactionResponseSubscription(subscription, stream)
}

func (t *TransactionService) AddClientPublicKey(_ context.Context, subscription *protos.TransactionResponseEventSubscription) (*protos.Empty, error) {
	_ = t.transactionProcessor.AddClientPublicKey(subscription)
	return &protos.Empty{}, nil
}

func (t *TransactionService) SubscribeNodeTransactions(subscription *protos.TransactionResponseEventSubscription,
	stream protos.TransactionService_SubscribeNodeTransactionsServer) error {
	return t.transactionProcessor.NodeTransactionResponseSubscriptionOrderlessChain(subscription, stream)
}

func (t *TransactionService) SubscribeBlocksSyncHotStuff(subscriptionEvent *protos.BlockEventSubscription,
	stream protos.TransactionService_SubscribeBlocksSyncHotStuffServer) error {
	return t.transactionProcessor.BlockSubscriptionFromOtherNodesSyncHotStuff(subscriptionEvent, stream)
}

func (t *TransactionService) StopAndGetProfilingResult(pr *protos.Profiling, respStream protos.TransactionService_StopAndGetProfilingResultServer) error {
	reportPath := logger.LogsPath
	if pr.ProfilingType == protos.Profiling_CPU {
		profiling.StopCPUProfiling()
		reportPath += "cpu.pprof"
	}
	if pr.ProfilingType == protos.Profiling_MEMORY {
		profiling.StopMemoryProfiling()
		reportPath += "mem.pprof"
	}

	profilingReport, err := os.Open(reportPath)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	defer func(report *os.File) {
		if err = report.Close(); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}(profilingReport)
	buffer := make([]byte, 64*1024)
	for {
		bytesRead, readErr := profilingReport.Read(buffer)
		if readErr != nil {
			if readErr != io.EOF {
				logger.ErrorLogger.Println(readErr)
			}
			break
		}
		response := &protos.ProfilingResult{
			Content: buffer[:bytesRead],
		}
		readErr = respStream.Send(response)
		if readErr != nil {
			logger.ErrorLogger.Println("Error while sending chunk:", readErr)
			return readErr
		}
	}
	return nil
}

func (t *TransactionService) GetTransactionProfilingResult(_ *protos.Empty, stream protos.TransactionService_GetTransactionProfilingResultServer) error {
	return t.transactionProcessor.SendTransactionProfiling(stream)
}

func (t *TransactionService) GetCPUMemoryProfilingResult(_ *protos.Empty, stream protos.TransactionService_GetCPUMemoryProfilingResultServer) error {
	return t.cpuMemoryProfiler.SendCPUMemoryProfilingNode(stream)
}
