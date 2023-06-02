package benchmark

import (
	"context"
	"encoding/csv"
	"github.com/google/uuid"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/connection/connpool"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ExperimentStatus struct {
	expectedDuration time.Duration
	clientName       string
	startTime        time.Time
	done             bool
	failed           bool
}

type Coordinator struct {
	BenchmarkName                    string
	Benchmark                        *Benchmark
	ClientPool                       map[string]*connpool.Pool
	NodePool                         map[string]*connpool.Pool
	OrdererPool                      map[string]*connpool.Pool
	SequencerPool                    map[string]*connpool.Pool
	inExperimentParticipatingNodes   []string
	inExperimentParticipatingClients []string
	ReportPath                       string
	ReportPathDetailed               string
	ReportProfilingPathDetailed      string
	ReportLatencyBreakDownDetailed   string
	ReportPathProfiling              string
	RoundCount                       int
	inFailureParticipatingNodes      []string
	inFailureParticipatingClients    []string
	transactionProfilingSummary      *TransactionProfilingStatSum
}

func NewCoordinator(benchmarkName string) *Coordinator {
	bc, _ := LoadBenchmark(benchmarkName)
	tempCoordinator := &Coordinator{
		BenchmarkName: benchmarkName,
		Benchmark:     bc,
		ClientPool:    connpool.GetAllClientsConnections(),
		NodePool:      connpool.GetAllNodesConnections(),
		OrdererPool:   connpool.GetAllOrderersConnections(),
		SequencerPool: connpool.GetAllSequencersConnections(),
		ReportPath: filepath.Join("./orderlesschain-experiments/results/",
			time.Now().Format("2006-01-02-15-04")+"-"+benchmarkName),
		ReportPathDetailed: filepath.Join("./orderlesschain-experiments/results/",
			time.Now().Format("2006-01-02-15-04")+"-"+benchmarkName, "details"),
		ReportProfilingPathDetailed: filepath.Join("./orderlesschain-experiments/results/",
			time.Now().Format("2006-01-02-15-04")+"-"+benchmarkName, "cpu-memory-details"),
		ReportPathProfiling: filepath.Join("./orderlesschain-experiments/results/",
			time.Now().Format("2006-01-02-15-04")+"-"+benchmarkName, "profiling"),
		ReportLatencyBreakDownDetailed: filepath.Join("./orderlesschain-experiments/results/",
			time.Now().Format("2006-01-02-15-04")+"-"+benchmarkName, "transaction-latency-details"),
		transactionProfilingSummary: initTransactionProfilingStatSum(),
	}
	sort.Strings(tempCoordinator.inExperimentParticipatingClients)
	sort.Strings(tempCoordinator.inExperimentParticipatingNodes)
	return tempCoordinator
}

func (b *Coordinator) CoordinateBenchmark() {
	b.changeModeRestart()
	time.Sleep(10 * time.Second)
	for _, round := range b.Benchmark.Rounds {
		b.setInExperimentParticipatingComponents(b.Benchmark, &round)
		b.RoundCount++
		round.ExperimentID = uuid.NewString()
		b.executeBenchmark(&round)
		if b.Benchmark.GossipIntervalMs > 0 {
			time.Sleep(time.Duration(b.Benchmark.TotalNodeCount*b.Benchmark.GossipIntervalMs) * time.Millisecond)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
	b.stopProfiling()
	b.makeTransactionProfiling()
	b.makeSummary()
	b.makeCPUMemoryProfiling()
	benchmarkFile, err := ioutil.ReadFile(filepath.Join(Path, b.BenchmarkName+".yml"))
	if err != nil {
		logger.ErrorLogger.Println(err)
	}
	err = ioutil.WriteFile(filepath.Join(b.ReportPath, "0config"+"-"+b.BenchmarkName+".yml"), benchmarkFile, os.ModePerm)
	if err != nil {
		logger.ErrorLogger.Println(err)
	}
}

func (b *Coordinator) changeModeRestart() {
	targetSystem, targetErr := GetTargetSystem(b.Benchmark.TargetSystem)
	if targetErr != nil {
		logger.FatalLogger.Fatalln(targetErr)
	}
	modeType := &protos.OperationMode{
		TargetSystem:                       targetSystem,
		Benchmark:                          b.BenchmarkName,
		GossipNodeCount:                    b.Benchmark.GossipNodeCount,
		TotalOrdererCount:                  b.Benchmark.TotalOrdererCount,
		TotalSequencerCount:                b.Benchmark.TotalSequencerCount,
		TotalNodeCount:                     b.Benchmark.TotalNodeCount,
		TotalClientCount:                   b.Benchmark.TotalClientCount,
		GossipIntervalMs:                   b.Benchmark.GossipIntervalMs,
		TransactionTimeoutSecond:           b.Benchmark.TransactionTimeoutSecond,
		BlockTimeOutMs:                     b.Benchmark.BlockTimeOutMs,
		BlockTransactionSize:               b.Benchmark.BlockTransactionSize,
		EndorsementPolicy:                  b.Benchmark.EndorsementPolicyOrgs,
		ProposalQueueConsumptionRateTps:    b.Benchmark.ProposalQueueConsumptionRateTPS,
		TransactionQueueConsumptionRateTps: b.Benchmark.TransactionQueueConsumptionRateTPS,
		QueueTickerDurationMs:              b.Benchmark.QueueTickerDurationMS,
		ExtraEndorsementOrgs:               b.Benchmark.ExtraEndorsementOrgs,
		OrgsPercentageIncreasedLoad:        b.Benchmark.OrgsPercentageIncreasedLoad,
		LoadIncreasePercentage:             b.Benchmark.LoadIncreasePercentage,
	}
	if modeType.TotalOrdererCount == 0 {
		modeType.TotalOrdererCount = int32(len(config.Config.Orderers))
	}
	if modeType.TotalSequencerCount == 0 {
		modeType.TotalSequencerCount = int32(len(config.Config.Sequencers))
	}
	if modeType.TotalNodeCount == 0 {
		modeType.TotalNodeCount = int32(len(config.Config.Nodes))
	}
	if modeType.TotalClientCount == 0 {
		modeType.TotalClientCount = int32(len(config.Config.Clients))
	}
	if modeType.GossipNodeCount == 0 {
		modeType.GossipNodeCount = modeType.TotalNodeCount
	}
	logger.InfoLogger.Println("Sending target system to components:", modeType.TargetSystem.String())
	wg := &sync.WaitGroup{}
	wg.Add(len(config.Config.Orderers))
	if strings.Contains(b.Benchmark.ProfilingComponents, "orderer") {
		modeType.ProfilingEnabled = b.Benchmark.Profiling
	} else {
		modeType.ProfilingEnabled = ""
	}
	for name := range config.Config.Orderers {
		go func(name string, wg *sync.WaitGroup) {
			conn, err := b.OrdererPool[name].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewOrdererServiceClient(conn.ClientConn)
			_, err = client.ChangeModeRestart(context.Background(), modeType)
			if err != nil {
				logger.ErrorLogger.Println(name, err)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(name, errCon)
			}
			wg.Done()
		}(name, wg)
	}
	wg.Wait()
	time.Sleep(5 * time.Second)
	wg = &sync.WaitGroup{}
	wg.Add(len(config.Config.Sequencers))
	if strings.Contains(b.Benchmark.ProfilingComponents, "sequencer") {
		modeType.ProfilingEnabled = b.Benchmark.Profiling
	} else {
		modeType.ProfilingEnabled = ""
	}
	for name := range config.Config.Sequencers {
		go func(name string, wg *sync.WaitGroup) {
			conn, err := b.SequencerPool[name].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewSequencerServiceClient(conn.ClientConn)
			_, err = client.ChangeModeRestart(context.Background(), modeType)
			if err != nil {
				logger.ErrorLogger.Println(name, err)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(name, errCon)
			}
			wg.Done()
		}(name, wg)
	}
	wg.Wait()
	time.Sleep(5 * time.Second)
	wg = &sync.WaitGroup{}
	wg.Add(len(config.Config.Nodes))
	if strings.Contains(b.Benchmark.ProfilingComponents, "node") {
		modeType.ProfilingEnabled = b.Benchmark.Profiling
	} else {
		modeType.ProfilingEnabled = ""
	}
	for name := range config.Config.Nodes {
		go func(name string, wg *sync.WaitGroup) {
			conn, err := b.NodePool[name].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewTransactionServiceClient(conn.ClientConn)
			_, err = client.ChangeModeRestart(context.Background(), modeType)
			if err != nil {
				logger.ErrorLogger.Println(name, err)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(name, errCon)
			}
			wg.Done()
		}(name, wg)
	}
	wg.Wait()
	time.Sleep(5 * time.Second)
	wg = &sync.WaitGroup{}
	wg.Add(len(config.Config.Clients))
	if strings.Contains(b.Benchmark.ProfilingComponents, "client") {
		modeType.ProfilingEnabled = b.Benchmark.Profiling
	} else {
		modeType.ProfilingEnabled = ""
	}
	for name := range config.Config.Clients {
		go func(name string, wg *sync.WaitGroup) {
			conn, err := b.ClientPool[name].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewBenchmarkServiceClient(conn.ClientConn)
			_, err = client.ChangeModeRestart(context.Background(), modeType)
			if err != nil {
				logger.ErrorLogger.Println(name, err)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(name, errCon)
			}
			wg.Done()
		}(name, wg)
	}
	wg.Wait()
	time.Sleep(5 * time.Second)
}

func (b *Coordinator) setInExperimentParticipatingComponents(benchmark *Benchmark, round *Round) {
	b.inExperimentParticipatingClients = []string{}
	for i := 0; i < round.NumberOfClients; i++ {
		b.inExperimentParticipatingClients = append(b.inExperimentParticipatingClients, "client"+strconv.Itoa(i))
	}
	b.inExperimentParticipatingNodes = []string{}
	for i := 0; i < int(benchmark.TotalNodeCount); i++ {
		b.inExperimentParticipatingNodes = append(b.inExperimentParticipatingNodes, "node"+strconv.Itoa(i))
	}
	if len(round.Failures) > 0 {
		b.inFailureParticipatingClients = []string{}
		for i := 0; i < len(b.inExperimentParticipatingClients); i++ {
			b.inFailureParticipatingClients = append(b.inFailureParticipatingClients, "client"+strconv.Itoa(i))
		}
		nodeFailureType := 1
		switch nodeFailureType {
		case 0:
			b.inFailureParticipatingNodes = []string{}
			for i := 0; i < len(b.inExperimentParticipatingNodes); i++ {
				b.inFailureParticipatingNodes = append(b.inFailureParticipatingNodes, "node"+strconv.Itoa(i))
			}
		case 1:
			failureRandom := rand.New(rand.NewSource(time.Now().UnixNano()))
			tempNodeUnique := make(map[string]bool)
			for i := 0; i < len(b.inExperimentParticipatingNodes); i++ {
				nodeIdString := "node" + strconv.Itoa(failureRandom.Intn(len(b.inExperimentParticipatingNodes)))
				if _, ok := tempNodeUnique[nodeIdString]; !ok {
					b.inFailureParticipatingNodes = append(b.inFailureParticipatingNodes, nodeIdString)
					tempNodeUnique[nodeIdString] = true
				} else {
					i--
				}
			}
		case 2:
			b.inFailureParticipatingNodes = []string{}
			endorsingOrgsCount := int(benchmark.EndorsementPolicyOrgs + benchmark.ExtraEndorsementOrgs)
			tempAllOrder := make(map[int][]string)
			maxSizeList := 0
			for currentClientIdInt := 0; currentClientIdInt < len(b.inExperimentParticipatingClients); currentClientIdInt++ {
				totalNodeConnectionsCounter := 0
				for clientCount := 0; clientCount < len(b.inExperimentParticipatingClients); clientCount++ {
					for endorser := 0; endorser < endorsingOrgsCount; endorser++ {
						nodeId := totalNodeConnectionsCounter % len(b.inExperimentParticipatingNodes)
						totalNodeConnectionsCounter++
						if clientCount == currentClientIdInt {
							tempAllOrder[currentClientIdInt] = append(tempAllOrder[currentClientIdInt], "node"+strconv.Itoa(nodeId))
							if maxSizeList < len(tempAllOrder[currentClientIdInt]) {
								maxSizeList = len(tempAllOrder[currentClientIdInt])
							}
						}
					}
				}
			}
			tempNodeUnique := make(map[string]bool)
			for nodeCounter := 0; nodeCounter < len(b.inExperimentParticipatingClients); nodeCounter++ {
				for i := 0; i < maxSizeList; i++ {
					theNodeAdded := false
					for currentClientIdInt := 0; currentClientIdInt < len(b.inExperimentParticipatingClients); currentClientIdInt++ {
						if i < len(tempAllOrder[currentClientIdInt]) {
							nodeIdString := tempAllOrder[currentClientIdInt][i]
							if _, ok := tempNodeUnique[nodeIdString]; !ok {
								b.inFailureParticipatingNodes = append(b.inFailureParticipatingNodes, nodeIdString)
								tempNodeUnique[nodeIdString] = true
								theNodeAdded = true
								break
							}
						}
					}
					if theNodeAdded {
						break
					}
				}
			}
		}

	}
}

func (b *Coordinator) executeBenchmark(round *Round) {
	logger.InfoLogger.Println("Round: ", round.Label)
	experimentStatuses := map[string]*ExperimentStatus{}
	wg := &sync.WaitGroup{}
	wg.Add(len(b.inExperimentParticipatingClients))
	for _, name := range b.inExperimentParticipatingClients {
		experimentStatuses[name] = &ExperimentStatus{
			clientName: name,
			startTime:  time.Now(),
		}
		bufferedMinutes := int(b.Benchmark.TransactionTimeoutSecond/60) + 1 // Buffer minutes for timing out the client experiment
		if bufferedMinutes == 0 {
			logger.FatalLogger.Println("buffered minutes cannot be zero")
		}
		if round.TransactionsSendDurationSecond > 0 {
			experimentStatuses[name].expectedDuration = time.Duration((round.TransactionsSendDurationSecond/60)+bufferedMinutes) * time.Minute
		} else {
			experimentStatuses[name].expectedDuration = time.Duration(((round.TotalTransactions/round.TotalSubmissionRate)/60)+bufferedMinutes) * time.Minute
		}
		go func(name string, wg *sync.WaitGroup) {
			conn, err := b.ClientPool[name].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewBenchmarkServiceClient(conn.ClientConn)
			targetSystem, err := GetTargetSystem(b.Benchmark.TargetSystem)
			if err != nil {
				logger.FatalLogger.Fatalln(err)
			}
			_, err = client.ExecuteBenchmark(context.Background(), &protos.BenchmarkConfig{
				Base: &protos.ExperimentBase{
					ExperimentId: round.ExperimentID,
				},
				ContractName:                  b.Benchmark.ContactName,
				TargetSystem:                  targetSystem,
				ReportImportance:              round.ReportImportance,
				BenchmarkFunctionName:         round.BenchmarkFunctionName,
				TransactionSendDurationSecond: int64(round.TransactionsSendDurationSecond),
				TotalTransactions:             int64(round.TotalTransactions / round.NumberOfClients),
				TransactionPerSecond:          int32(round.TotalSubmissionRate / round.NumberOfClients),
				NumberOfKeys:                  int64(round.NumberOfKeys),
				NumberOfKeysSecond:            int64(round.NumberOfKeysSecond),
				NumberOfKeysThird:             int64(round.NumberOfKeysThird),
				EndorsementPolicyOrgs:         b.Benchmark.EndorsementPolicyOrgs,
				CrdtObjectCount:               round.CrdtObjectCount,
				CrdtOperationPerObjectCount:   round.CrdtOperationPerObjectCount,
				CrdtObjectType:                round.CrdtObjectType,
			})
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(errCon)
			}
			wg.Done()
		}(name, wg)
	}
	wg.Wait()
	if len(round.Failures) > 0 {
		go b.runFailures(round)
	}
	monitorTimer := time.NewTicker(5 * time.Second)
	for {
		doneCounter := 0
		for _, experimentStatus := range experimentStatuses {
			if experimentStatus.done {
				doneCounter++
			}
		}
		if len(experimentStatuses) == doneCounter {
			break
		}
		<-monitorTimer.C
		for _, experimentStatus := range experimentStatuses {
			if !experimentStatus.done {
				conn, err := b.ClientPool[experimentStatus.clientName].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
					continue
				}
				client := protos.NewBenchmarkServiceClient(conn.ClientConn)
				result, err := client.ExecutionStatus(context.Background(), &protos.ExperimentBase{
					ExperimentId: round.ExperimentID,
				})
				if errCon := conn.Close(); errCon != nil {
					logger.ErrorLogger.Println(errCon)
				}
				if err != nil {
					logger.InfoLogger.Println(experimentStatus.clientName)
					logger.ErrorLogger.Println(err)
					continue
				}
				if result.ExperimentStatus == protos.ExperimentResult_DONE {
					experimentStatus.done = true
				}
				if result.ExperimentStatus == protos.ExperimentResult_FAILED {
					logger.ErrorLogger.Println("Experiment failed at", experimentStatus.clientName)
					experimentStatus.done = true
					experimentStatus.failed = true
				}
				executionDuration := time.Since(experimentStatus.startTime)
				if executionDuration > experimentStatus.expectedDuration {
					logger.ErrorLogger.Println("Experiment timed out at", experimentStatus.clientName)
					experimentStatus.done = true
					experimentStatus.failed = true
				}
			}
		}

	}
	if err := os.MkdirAll(b.ReportPath, os.ModePerm); err != nil {
		logger.ErrorLogger.Println(err)
	}
	if err := os.MkdirAll(b.ReportPathDetailed, os.ModePerm); err != nil {
		logger.ErrorLogger.Println(err)
	}
	for _, experimentStatus := range experimentStatuses {
		if !experimentStatus.failed {
			conn, err := b.ClientPool[experimentStatus.clientName].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewBenchmarkServiceClient(conn.ClientConn)
			_, err = client.ExecutionStatus(context.Background(), &protos.ExperimentBase{
				ExperimentId: round.ExperimentID,
			})
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
			reportStream, err := client.GetExperimentResult(context.Background(), &protos.ExperimentBase{
				ExperimentId: round.ExperimentID,
			})
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
			var reportData []byte
			for {
				dataChunk, err := reportStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					logger.ErrorLogger.Println(err)
					break
				}
				reportData = append(reportData, dataChunk.Content...)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(errCon)
			}
			if len(reportData) > 0 {
				err = ioutil.WriteFile(filepath.Join(b.ReportPathDetailed, strconv.Itoa(b.RoundCount)+"-"+round.BenchmarkFunctionName+"-"+experimentStatus.clientName+".csv"), reportData, os.ModePerm)
				if err != nil {
					logger.ErrorLogger.Println(err)
				}
			}
		}
	}
}

func (b *Coordinator) stopProfiling() {
	if len(b.Benchmark.Profiling) == 0 || b.Benchmark.Profiling == "enable_server" || b.Benchmark.Profiling == "bandwidth" {
		return
	}
	var profilingType protos.Profiling_ProfilingType
	var profilingFileName string
	if b.Benchmark.Profiling == "cpu" {
		profilingType = protos.Profiling_CPU
		profilingFileName = "cpu"
	} else if b.Benchmark.Profiling == "memory" {
		profilingType = protos.Profiling_MEMORY
		profilingFileName = "memo"
	} else {
		logger.ErrorLogger.Println("invalid profiling type")
		return
	}
	err := os.MkdirAll(b.ReportPathProfiling, os.ModePerm)
	if err != nil {
		logger.ErrorLogger.Println(err)
	}

	if strings.Contains(b.Benchmark.ProfilingComponents, "orderer") {
		for ordererName := range config.Config.Orderers {
			conn, err := b.OrdererPool[ordererName].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewOrdererServiceClient(conn.ClientConn)
			reportStream, err := client.StopAndGetProfilingResult(context.Background(), &protos.Profiling{
				ProfilingType: profilingType,
			})
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
			var reportData []byte
			for {
				dataChunk, err := reportStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					logger.ErrorLogger.Println(err)
					break
				}
				reportData = append(reportData, dataChunk.Content...)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(errCon)
			}
			if len(reportData) > 0 {
				err = ioutil.WriteFile(filepath.Join(b.ReportPathProfiling, profilingFileName+"-"+ordererName+".pprof"), reportData, os.ModePerm)
				if err != nil {
					logger.ErrorLogger.Println(err)
				}
			}
		}
	}

	if strings.Contains(b.Benchmark.ProfilingComponents, "node") {
		for _, nodeName := range b.inExperimentParticipatingNodes {
			conn, err := b.NodePool[nodeName].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewTransactionServiceClient(conn.ClientConn)
			reportStream, err := client.StopAndGetProfilingResult(context.Background(), &protos.Profiling{
				ProfilingType: profilingType,
			})
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
			var reportData []byte
			for {
				dataChunk, err := reportStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					logger.ErrorLogger.Println(err)
					break
				}
				reportData = append(reportData, dataChunk.Content...)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(errCon)
			}
			if len(reportData) > 0 {
				err = ioutil.WriteFile(filepath.Join(b.ReportPathProfiling, profilingFileName+"-"+nodeName+".pprof"), reportData, os.ModePerm)
				if err != nil {
					logger.ErrorLogger.Println(err)
				}
			}
		}
	}

	if strings.Contains(b.Benchmark.ProfilingComponents, "client") {
		for _, clientName := range b.inExperimentParticipatingClients {
			conn, err := b.ClientPool[clientName].Get(context.Background())
			if conn == nil || err != nil {
				logger.ErrorLogger.Println(err)
			}
			client := protos.NewBenchmarkServiceClient(conn.ClientConn)
			reportStream, err := client.StopAndGetProfilingResult(context.Background(), &protos.Profiling{
				ProfilingType: profilingType,
			})
			if err != nil {
				logger.ErrorLogger.Println(err)
			}
			var reportData []byte
			for {
				dataChunk, err := reportStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					logger.ErrorLogger.Println(err)
					break
				}
				reportData = append(reportData, dataChunk.Content...)
			}
			if errCon := conn.Close(); errCon != nil {
				logger.ErrorLogger.Println(errCon)
			}
			if len(reportData) > 0 {
				err = ioutil.WriteFile(filepath.Join(b.ReportPathProfiling, profilingFileName+"-"+clientName+".pprof"), reportData, os.ModePerm)
				if err != nil {
					logger.ErrorLogger.Println(err)
				}
			}
		}
	}
}

func (b *Coordinator) runFailures(round *Round) {
	for _, failure := range round.Failures {
		go b.runSingleFailures(failure)
	}
}

func (b *Coordinator) runSingleFailures(failure Failure) {
	var failureType protos.FailureType
	switch failure.FailureType {
	case "crashed":
		failureType = protos.FailureType_CRASHED
	case "tampered":
		failureType = protos.FailureType_TAMPERED
	case "not_responding":
		failureType = protos.FailureType_NOTRESPONDING
	case "random":
		failureType = protos.FailureType_RANODM
	default:
		logger.ErrorLogger.Println("No Failure Type")
		return
	}
	timerWaitingForFailure := time.NewTimer(time.Duration(failure.FailureStartS) * time.Second)
	<-timerWaitingForFailure.C

	var sentFailureToNodes []string
	if failure.FailedOrgsCount > 0 {
		wgNodes := &sync.WaitGroup{}
		wgNodes.Add(int(failure.FailedOrgsCount))
		counter := int32(0)
		for _, name := range b.inFailureParticipatingNodes {
			counter++
			if counter > failure.FailedOrgsCount {
				break
			}
			sentFailureToNodes = append(sentFailureToNodes, name)
			go func(name string, thisFailure Failure, failType protos.FailureType, wg *sync.WaitGroup) {
				conn, err := b.NodePool[name].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
				}
				client := protos.NewTransactionServiceClient(conn.ClientConn)
				_, err = client.FailureCommand(context.Background(), &protos.FailureCommandMode{
					FailureDurationS: thisFailure.DurationS,
					FailureType:      failType,
				})
				if err != nil {
					logger.ErrorLogger.Println(name, err)
				}
				if errCon := conn.Close(); errCon != nil {
					logger.ErrorLogger.Println(name, errCon)
				}
				wg.Done()
			}(name, failure, failureType, wgNodes)
		}
		wgNodes.Wait()
	}
	if failure.FailedClientPercentage > 0 {
		failedClientsLength := int(float32(len(b.inFailureParticipatingClients)) *
			(float32(failure.FailedClientPercentage) / 100.00))
		logger.InfoLogger.Println("clients in failure ", failedClientsLength)
		wgClientFailures := &sync.WaitGroup{}
		wgClientFailures.Add(failedClientsLength)
		failedCounter := 0
		for _, name := range b.inFailureParticipatingClients {
			failedCounter++
			if failedCounter > failedClientsLength {
				break
			}
			go func(name string, thisFailure Failure, failType protos.FailureType, wg *sync.WaitGroup) {
				conn, err := b.ClientPool[name].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
				}
				client := protos.NewBenchmarkServiceClient(conn.ClientConn)
				_, err = client.FailureCommand(context.Background(), &protos.FailureCommandMode{
					FailureDurationS: thisFailure.DurationS,
					FailureType:      failType,
				})
				if err != nil {
					logger.ErrorLogger.Println(name, err)
				}
				if errCon := conn.Close(); errCon != nil {
					logger.ErrorLogger.Println(name, errCon)
				}
				wg.Done()
			}(name, failure, failureType, wgClientFailures)
		}
		wgClientFailures.Wait()
	}

	if failure.NotifyClients && failure.FailedOrgsCount > 0 {
		wgClients := &sync.WaitGroup{}
		wgClients.Add(len(b.inExperimentParticipatingClients))
		for _, name := range b.inExperimentParticipatingClients {
			go func(name string, wg *sync.WaitGroup) {
				conn, err := b.ClientPool[name].Get(context.Background())
				if conn == nil || err != nil {
					logger.ErrorLogger.Println(err)
				}
				client := protos.NewBenchmarkServiceClient(conn.ClientConn)
				_, err = client.FaultyNodesNotify(context.Background(), &protos.FaultyNodes{
					NodeId:      sentFailureToNodes,
					StartAfterS: failure.ClientStartAfterS,
				})
				if err != nil {
					logger.ErrorLogger.Println(name, err)
				}
				if errCon := conn.Close(); errCon != nil {
					logger.ErrorLogger.Println(name, errCon)
				}
				wg.Done()
			}(name, wgClients)
		}
		wgClients.Wait()
	}
}

func (b *Coordinator) makeSummary() {
	summary := CalculateExperimentsStats(b.ReportPathDetailed)
	summaryPath := filepath.Join(b.ReportPath, "0-summary.csv")
	summaryFile, err := os.Create(summaryPath)
	if err != nil {
		logger.ErrorLogger.Println("failed creating file:", err)
	}
	csvSummaryWriter := csv.NewWriter(summaryFile)
	_ = csvSummaryWriter.Write([]string{
		"Round",
		"TX Count",
		"Success TX Count",
		"Fail TX Count",
		"Submission TSP",
		"Throughput TSP",
		"Max Latency MS",
		"Min Latency MS",
		"Average Latency MS",
		"01th PCTL MS",
		"05th PCTL MS",
		"10th PCTL MS",
		"25th PCTL MS",
		"50th PCTL MS",
		"75th PCTL MS",
		"90th PCTL MS",
		"95th PCTL MS",
		"99th PCTL MS",
		"Total Sent KB",
		"Total Sent KB/S",
		"Total Received KB",
		"Total Received KB/S",
	})
	_ = csvSummaryWriter.Write([]string{
		"Summary",
		strconv.Itoa(summary.totalTransactionsALL),
		strconv.Itoa(summary.succeededTransactionsALL),
		strconv.Itoa(summary.failedTransactionsALL),
		strconv.Itoa(summary.submitRatePerSecALL),
		strconv.Itoa(summary.throughputPerSecALL),
		strconv.FormatInt(summary.maxLatencyMsALL, 10),
		strconv.FormatInt(summary.minLatencyMsALL, 10),
		strconv.FormatInt(summary.averageLatencyMsALL, 10),
		strconv.FormatInt(summary.percentile01LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile05LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile10LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile25LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile50LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile75LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile90LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile95LatencyMsALL, 10),
		strconv.FormatInt(summary.percentile99LatencyMsALL, 10),
		strconv.Itoa(summary.TotalSentKBALL),
		strconv.Itoa(summary.TotalSentPerSecondKBALL),
		strconv.Itoa(summary.TotalReceivedKBALL),
		strconv.Itoa(summary.TotalReceivedPerSecondKBALL),
	})
	_ = csvSummaryWriter.Write([]string{
		"Summary Write",
		strconv.Itoa(summary.totalTransactionsWrite),
		strconv.Itoa(summary.succeededTransactionsWrite),
		strconv.Itoa(summary.failedTransactionsWrite),
		strconv.Itoa(summary.submitRatePerSecWrite),
		strconv.Itoa(summary.throughputPerSecWrite),
		strconv.FormatInt(summary.maxLatencyMsWrite, 10),
		strconv.FormatInt(summary.minLatencyMsWrite, 10),
		strconv.FormatInt(summary.averageLatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile01LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile05LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile10LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile25LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile50LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile75LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile90LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile95LatencyMsWrite, 10),
		strconv.FormatInt(summary.percentile99LatencyMsWrite, 10),
		"0",
		"0",
		"0",
		"0",
	})
	_ = csvSummaryWriter.Write([]string{
		"Summary Read",
		strconv.Itoa(summary.totalTransactionsRead),
		strconv.Itoa(summary.succeededTransactionsRead),
		strconv.Itoa(summary.failedTransactionsRead),
		strconv.Itoa(summary.submitRatePerSecRead),
		strconv.Itoa(summary.throughputPerSecRead),
		strconv.FormatInt(summary.maxLatencyMsRead, 10),
		strconv.FormatInt(summary.minLatencyMsRead, 10),
		strconv.FormatInt(summary.averageLatencyMsRead, 10),
		strconv.FormatInt(summary.percentile01LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile05LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile10LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile25LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile50LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile75LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile90LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile95LatencyMsRead, 10),
		strconv.FormatInt(summary.percentile99LatencyMsRead, 10),
		"0",
		"0",
		"0",
		"0",
	})
	for clientName, clientStat := range summary.clientStats {
		_ = csvSummaryWriter.Write([]string{
			clientName,
			strconv.Itoa(len(clientStat.transactionsALL)),
			strconv.Itoa(clientStat.succeededTransactionsALL),
			strconv.Itoa(clientStat.failedTransactionsALL),
			strconv.Itoa(clientStat.submitRatePerSecALL),
			strconv.Itoa(clientStat.throughputPerSecALL),
			strconv.FormatInt(clientStat.maxLatencyMsALL, 10),
			strconv.FormatInt(clientStat.minLatencyMsALL, 10),
			strconv.FormatInt(clientStat.averageLatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile01LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile05LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile10LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile25LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile50LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile75LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile90LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile95LatencyMsALL, 10),
			strconv.FormatInt(clientStat.percentile99LatencyMsALL, 10),
			strconv.Itoa(clientStat.TotalSentKBALL),
			strconv.Itoa(clientStat.TotalSentPerSecondKBALL),
			strconv.Itoa(clientStat.TotalReceivedKBALL),
			strconv.Itoa(clientStat.TotalReceivedPerSecondKBALL),
		})
	}
	csvSummaryWriter.Flush()
	logger.InfoLogger.Println("Summary report file generated")
	if err = summaryFile.Close(); err != nil {
		logger.ErrorLogger.Println(err)
	}
}

func (b *Coordinator) makeTransactionProfiling() {
	if !config.Config.IsTransactionProfilingEnabled {
		return
	}
	if err := os.MkdirAll(b.ReportLatencyBreakDownDetailed, os.ModePerm); err != nil {
		logger.ErrorLogger.Println(err)
	}
	for name := range config.Config.Orderers {
		conn, err := b.OrdererPool[name].Get(context.Background())
		if conn == nil || err != nil {
			logger.ErrorLogger.Println(err)
		}
		client := protos.NewOrdererServiceClient(conn.ClientConn)
		stream, err := client.GetTransactionProfilingResult(context.Background(), &protos.Empty{})
		for {
			txProfile, err := stream.Recv()
			if err == io.EOF || err != nil {
				break
			}
			b.summarizeTransactionProfiling(txProfile)
		}
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(name, errCon)
		}
	}
	for name := range config.Config.Sequencers {
		conn, err := b.SequencerPool[name].Get(context.Background())
		if conn == nil || err != nil {
			logger.ErrorLogger.Println(err)
		}
		client := protos.NewSequencerServiceClient(conn.ClientConn)
		stream, err := client.GetTransactionProfilingResult(context.Background(), &protos.Empty{})
		for {
			txProfile, err := stream.Recv()
			if err == io.EOF || err != nil {
				break
			}
			b.summarizeTransactionProfiling(txProfile)
		}
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(name, errCon)
		}
	}
	for _, name := range b.inExperimentParticipatingNodes {
		conn, err := b.NodePool[name].Get(context.Background())
		if conn == nil || err != nil {
			logger.ErrorLogger.Println(err)
		}
		client := protos.NewTransactionServiceClient(conn.ClientConn)
		stream, err := client.GetTransactionProfilingResult(context.Background(), &protos.Empty{})
		for {
			txProfile, err := stream.Recv()
			if err == io.EOF || err != nil {
				break
			}
			b.summarizeTransactionProfiling(txProfile)
		}
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(name, errCon)
		}
	}
	b.finalizeTransactionProfiling()
}

func (b *Coordinator) summarizeTransactionProfiling(txProfile *protos.LatencyBreakDown) {
	txStat := &TransactionProfilingStat{}
	var ok bool
	if txStat, ok = b.transactionProfilingSummary.transactionProfiles[txProfile.TransactionId]; !ok {
		b.transactionProfilingSummary.transactionProfiles[txProfile.TransactionId] = &TransactionProfilingStat{
			transactionId: txProfile.TransactionId,
		}
		txStat = b.transactionProfilingSummary.transactionProfiles[txProfile.TransactionId]
	}
	if txProfile.SequenceDuration > 0 {
		txStat.sequenceDuration = txProfile.SequenceDuration
	}
	if txProfile.OrderDuration > 0 {
		txStat.orderDuration = txProfile.OrderDuration
	}
	if txProfile.EndorseDuration > 0 {
		txStat.endorseDuration += txProfile.EndorseDuration
		txStat.nodeCountEndorse++
	}
	if txProfile.ExecBidlDuration > 0 {
		txStat.execDuration += txProfile.ExecBidlDuration
		txStat.nodeCountExec++
	}
	if txProfile.CommitDuration > 0 {
		txStat.commitDuration += txProfile.CommitDuration
		txStat.nodeCountCommit++
	}
}

func (b *Coordinator) finalizeTransactionProfiling() {
	if len(b.transactionProfilingSummary.transactionProfiles) == 0 {
		return
	}
	for _, txStat := range b.transactionProfilingSummary.transactionProfiles {
		if txStat.nodeCountEndorse > 0 {
			txStat.endorseDurationAverage = txStat.endorseDuration / int64(txStat.nodeCountEndorse)
		}
		if txStat.nodeCountExec > 0 {
			txStat.execDurationAverage = txStat.execDuration / int64(txStat.nodeCountExec)
		}
		if txStat.nodeCountCommit > 0 {
			txStat.commitDurationAverage = txStat.commitDuration / int64(txStat.nodeCountCommit)
		}
	}
	totalPath := filepath.Join(b.ReportLatencyBreakDownDetailed, "0-transaction-latency-profiling.csv")
	totalFile, err := os.Create(totalPath)
	if err != nil {
		logger.ErrorLogger.Println("failed creating file:", err)
	}
	csvWriter := csv.NewWriter(totalFile)
	_ = csvWriter.Write([]string{
		"TX ID",
		"Sequence Duration MicroS",
		"Order Duration MicroS",
		"Endorse Duration Average MicroS",
		"Execute Duration Average MicroS",
		"Commit Duration Average MicroS",
	})
	for _, txStat := range b.transactionProfilingSummary.transactionProfiles {
		_ = csvWriter.Write([]string{
			txStat.transactionId,
			strconv.FormatInt(txStat.sequenceDuration/int64(time.Microsecond), 10),
			strconv.FormatInt(txStat.orderDuration/int64(time.Microsecond), 10),
			strconv.FormatInt(txStat.endorseDurationAverage/int64(time.Microsecond), 10),
			strconv.FormatInt(txStat.execDurationAverage/int64(time.Microsecond), 10),
			strconv.FormatInt(txStat.commitDurationAverage/int64(time.Microsecond), 10),
		})
	}
	csvWriter.Flush()
	if err = totalFile.Close(); err != nil {
		logger.ErrorLogger.Println(err)
	}

	summaryStat := &TransactionProfilingStat{}
	for _, txStat := range b.transactionProfilingSummary.transactionProfiles {
		summaryStat.sequenceDuration += txStat.sequenceDuration
		summaryStat.orderDuration += txStat.orderDuration
		summaryStat.endorseDurationAverage += txStat.endorseDurationAverage
		summaryStat.execDurationAverage += txStat.execDurationAverage
		summaryStat.commitDurationAverage += txStat.commitDurationAverage
	}

	transactionsLength := int64(len(b.transactionProfilingSummary.transactionProfiles))
	summaryStat.sequenceDuration = summaryStat.sequenceDuration / transactionsLength
	summaryStat.orderDuration = summaryStat.orderDuration / transactionsLength
	summaryStat.endorseDurationAverage = summaryStat.endorseDurationAverage / transactionsLength
	summaryStat.execDurationAverage = summaryStat.execDurationAverage / transactionsLength
	summaryStat.commitDurationAverage = summaryStat.commitDurationAverage / transactionsLength

	summaryPath := filepath.Join(b.ReportPath, "0-latency-break-down-summary.csv")
	summaryFile, err := os.Create(summaryPath)
	if err != nil {
		logger.ErrorLogger.Println("failed creating file:", err)
	}
	csvWriter = csv.NewWriter(summaryFile)
	_ = csvWriter.Write([]string{
		"Sequence Duration MicroS",
		"Order Duration MicroS",
		"Endorse Duration Average MicroS",
		"Execute Duration Average MicroS",
		"Commit Duration Average MicroS",
	})
	_ = csvWriter.Write([]string{
		strconv.FormatInt(summaryStat.sequenceDuration/int64(time.Microsecond), 10),
		strconv.FormatInt(summaryStat.orderDuration/int64(time.Microsecond), 10),
		strconv.FormatInt(summaryStat.endorseDurationAverage/int64(time.Microsecond), 10),
		strconv.FormatInt(summaryStat.execDurationAverage/int64(time.Microsecond), 10),
		strconv.FormatInt(summaryStat.commitDurationAverage/int64(time.Microsecond), 10),
	})
	csvWriter.Flush()
	logger.InfoLogger.Println("Summary transaction profiling file generated")
	if err = summaryFile.Close(); err != nil {
		logger.ErrorLogger.Println(err)
	}
}

func (b *Coordinator) makeCPUMemoryProfiling() {
	if !config.Config.IsRealTimeCPUMemoryProfilingEnabled {
		return
	}
	if err := os.MkdirAll(b.ReportProfilingPathDetailed, os.ModePerm); err != nil {
		logger.ErrorLogger.Println(err)
	}
	CPUMemoryRecords := map[string]*protos.MemoryCPUProfile{}
	for name := range config.Config.Orderers {
		conn, err := b.OrdererPool[name].Get(context.Background())
		if conn == nil || err != nil {
			logger.ErrorLogger.Println(err)
		}
		client := protos.NewOrdererServiceClient(conn.ClientConn)
		stream, err := client.GetCPUMemoryProfilingResult(context.Background(), &protos.Empty{})
		for {
			profile, err := stream.Recv()
			if err == io.EOF || err != nil {
				break
			}
			CPUMemoryRecords[profile.Timestamp] = profile
		}
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(name, errCon)
		}
		b.writeCPUMemoryProfiling(name, CPUMemoryRecords)
	}
	CPUMemoryRecords = map[string]*protos.MemoryCPUProfile{}
	for name := range config.Config.Sequencers {
		conn, err := b.SequencerPool[name].Get(context.Background())
		if conn == nil || err != nil {
			logger.ErrorLogger.Println(err)
		}
		client := protos.NewSequencerServiceClient(conn.ClientConn)
		stream, err := client.GetCPUMemoryProfilingResult(context.Background(), &protos.Empty{})
		for {
			profile, err := stream.Recv()
			if err == io.EOF || err != nil {
				break
			}
			CPUMemoryRecords[profile.Timestamp] = profile
		}
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(name, errCon)
		}
		b.writeCPUMemoryProfiling(name, CPUMemoryRecords)
	}
	CPUMemoryRecords = map[string]*protos.MemoryCPUProfile{}
	for _, name := range b.inExperimentParticipatingNodes {
		conn, err := b.NodePool[name].Get(context.Background())
		if conn == nil || err != nil {
			logger.ErrorLogger.Println(err)
		}
		client := protos.NewTransactionServiceClient(conn.ClientConn)
		stream, err := client.GetCPUMemoryProfilingResult(context.Background(), &protos.Empty{})
		for {
			profile, err := stream.Recv()
			if err == io.EOF || err != nil {
				break
			}
			CPUMemoryRecords[profile.Timestamp] = profile
		}
		if errCon := conn.Close(); errCon != nil {
			logger.ErrorLogger.Println(name, errCon)
		}
		b.writeCPUMemoryProfiling(name, CPUMemoryRecords)
	}
}

func (b *Coordinator) writeCPUMemoryProfiling(name string, records map[string]*protos.MemoryCPUProfile) {
	profilingPath := filepath.Join(b.ReportProfilingPathDetailed, name+".csv")
	profilingFile, err := os.Create(profilingPath)
	if err != nil {
		logger.ErrorLogger.Println("failed creating file:", err)
	}
	csvWriter := csv.NewWriter(profilingFile)
	_ = csvWriter.Write([]string{
		"Timestamp",
		"Cpu Usage Percentage",
		"Allocated Heap Mb",
		"Heap In Use Mb",
		"Total Allocated Mb",
		"System Memory Mb",
	})
	for _, record := range records {
		_ = csvWriter.Write([]string{
			record.Timestamp,
			strconv.FormatInt(record.CpuUsagePercentage, 10),
			strconv.FormatInt(record.CpuUsagePercentage, 10),
			strconv.FormatInt(record.AllocatedHeap, 10),
			strconv.FormatInt(record.TotalAllocated, 10),
			strconv.FormatInt(record.SystemMemory, 10),
		})
	}
	csvWriter.Flush()
	if err = profilingFile.Close(); err != nil {
		logger.ErrorLogger.Println(err)
	}

}
