package sequencerservice

import (
	"context"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/cpumemoryprofiling"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"io"
)

type SequencerService struct {
	transactionProcessor *transactionprocessor.SequencerProcessor
	cpuMemoryProfiler    *cpumemoryprofiling.CPUMemoryProfiler
}

func NewSequencerService() *SequencerService {
	return &SequencerService{
		transactionProcessor: transactionprocessor.InitTransactionSequencerProcessor(),
		cpuMemoryProfiler:    cpumemoryprofiling.InitCPUMemoryProfiler(),
	}
}

func (s *SequencerService) BIDLTransactions(stream protos.SequencerService_BIDLTransactionsServer) error {
	for {
		transaction, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.Empty{})
		}
		if err != nil {
			return err
		}
		s.transactionProcessor.ProcessTransactionStream(transaction)
	}
}

func (s *SequencerService) SubscribeTransactionsForOrdering(subscriptionEvent *protos.SequencedTransactionForOrdering,
	stream protos.SequencerService_SubscribeTransactionsForOrderingServer) error {
	return s.transactionProcessor.SubscribeForOrdering(subscriptionEvent, stream)
}

func (s *SequencerService) SubscribeTransactionsForProcessing(subscriptionEvent *protos.SequencedTransactionForCommitting,
	stream protos.SequencerService_SubscribeTransactionsForProcessingServer) error {
	return s.transactionProcessor.SubscribeForCommitting(subscriptionEvent, stream)
}

func (s *SequencerService) ChangeModeRestart(_ context.Context, opm *protos.OperationMode) (*protos.Empty, error) {
	go config.UpdateModeAndRestart(opm)
	return &protos.Empty{}, nil
}

func (s *SequencerService) GetTransactionProfilingResult(_ *protos.Empty, stream protos.SequencerService_GetTransactionProfilingResultServer) error {
	return s.transactionProcessor.SendTransactionProfiling(stream)
}

func (s *SequencerService) GetCPUMemoryProfilingResult(_ *protos.Empty, stream protos.SequencerService_GetCPUMemoryProfilingResultServer) error {
	return s.cpuMemoryProfiler.SendCPUMemoryProfilingSequencer(stream)
}
