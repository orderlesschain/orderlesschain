package ordererservice

import (
	"context"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/cpumemoryprofiling"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/profiling"
	"gitlab.lrz.de/orderless/orderlesschain/internal/transactionprocessor"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"io"
	"os"
)

type OrdererService struct {
	transactionProcessor *transactionprocessor.OrdererProcessor
	cpuMemoryProfiler    *cpumemoryprofiling.CPUMemoryProfiler
}

func NewOrdererService() *OrdererService {
	return &OrdererService{
		transactionProcessor: transactionprocessor.InitTransactionOrdererProcessor(),
		cpuMemoryProfiler:    cpumemoryprofiling.InitCPUMemoryProfiler(),
	}
}

func (o *OrdererService) CommitFabricAndFabricCRDTTransactionStream(stream protos.OrdererService_CommitFabricAndFabricCRDTTransactionStreamServer) error {
	for {
		transaction, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.Empty{})
		}
		if err != nil {
			return err
		}
		o.transactionProcessor.ProcessTransactionStream(transaction)
	}
}

func (o *OrdererService) CommitSyncHotStuffTransactionStream(stream protos.OrdererService_CommitSyncHotStuffTransactionStreamServer) error {
	for {
		transaction, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&protos.Empty{})
		}
		if err != nil {
			return err
		}
		o.transactionProcessor.ProcessTransactionStream(transaction)
	}
}

func (o *OrdererService) SubscribeBlocks(subscriptionEvent *protos.BlockEventSubscription,
	stream protos.OrdererService_SubscribeBlocksServer) error {
	return o.transactionProcessor.BlockSubscription(subscriptionEvent, stream)
}

func (o *OrdererService) ChangeModeRestart(_ context.Context, opm *protos.OperationMode) (*protos.Empty, error) {
	go config.UpdateModeAndRestart(opm)
	return &protos.Empty{}, nil
}

func (o *OrdererService) StopAndGetProfilingResult(pr *protos.Profiling, respStream protos.OrdererService_StopAndGetProfilingResultServer) error {
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

func (o *OrdererService) GetTransactionProfilingResult(_ *protos.Empty, stream protos.OrdererService_GetTransactionProfilingResultServer) error {
	return o.transactionProcessor.SendTransactionProfiling(stream)
}

func (o *OrdererService) GetCPUMemoryProfilingResult(_ *protos.Empty, stream protos.OrdererService_GetCPUMemoryProfilingResultServer) error {
	return o.cpuMemoryProfiler.SendCPUMemoryProfilingOrderer(stream)
}
