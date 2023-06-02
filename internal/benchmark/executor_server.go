package benchmark

import (
	"context"
	"gitlab.lrz.de/orderless/orderlesschain/internal/benchmark/bencmarkdb"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/profiling"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"io"
	"os"
)

type ExecutorServer struct {
	executor *Executor
}

func NewExecutorServer() *ExecutorServer {
	return &ExecutorServer{
		executor: NewExecutor(),
	}
}

func (es *ExecutorServer) ExecuteBenchmark(_ context.Context, bconfig *protos.BenchmarkConfig) (*protos.Empty, error) {
	rex := NewRoundExecutor(bconfig, es.executor)
	go func(rex *RoundExecutor) {
		go rex.TransactionsDoneChecker()
		go rex.runExecuteBenchmarkStarter()
		rex.ExecuteBenchmark()
	}(rex)
	return &protos.Empty{}, nil
}

func (es *ExecutorServer) FaultyNodesNotify(_ context.Context, nodes *protos.FaultyNodes) (*protos.Empty, error) {
	go es.executor.setInExperimentParticipatingFaultyComponents(nodes)
	return &protos.Empty{}, nil
}

func (es *ExecutorServer) ExecutionStatus(_ context.Context, exp *protos.ExperimentBase) (*protos.ExperimentResult, error) {
	status, err := es.executor.dbOps.GetExperimentStatusWithExperimentID(exp.ExperimentId)
	if err != nil {
		return &protos.ExperimentResult{
			ExperimentStatus: protos.ExperimentResult_ExperimentStatus(bencmarkdb.Running),
		}, nil
	}
	return &protos.ExperimentResult{
		ExperimentStatus: protos.ExperimentResult_ExperimentStatus(status),
	}, nil
}

func (es *ExecutorServer) GetExperimentResult(exp *protos.ExperimentBase, respStream protos.BenchmarkService_GetExperimentResultServer) error {
	reportPath, err := es.executor.dbOps.GetExperimentResultPathWithExperimentID(exp.ExperimentId)
	if err != nil || len(reportPath) == 0 {
		logger.ErrorLogger.Println("Could not get the report path")
		return err
	}
	report, err := os.Open(reportPath)
	if err != nil {
		logger.ErrorLogger.Println(err)
		return err
	}
	defer func(report *os.File) {
		if err = report.Close(); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}(report)
	buffer := make([]byte, 64*1024)
	for {
		bytesRead, readErr := report.Read(buffer)
		if readErr != nil {
			if readErr != io.EOF {
				logger.ErrorLogger.Println(readErr)
			}
			break
		}
		response := &protos.ReportFile{
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

func (es *ExecutorServer) ChangeModeRestart(_ context.Context, opm *protos.OperationMode) (*protos.Empty, error) {
	go config.UpdateModeAndRestart(opm)
	return &protos.Empty{}, nil
}

func (es *ExecutorServer) StopAndGetProfilingResult(pr *protos.Profiling, respStream protos.BenchmarkService_StopAndGetProfilingResultServer) error {
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

func (es *ExecutorServer) FailureCommand(_ context.Context, fc *protos.FailureCommandMode) (*protos.Empty, error) {
	go es.executor.SetFailureCommand(fc)
	return &protos.Empty{}, nil
}
