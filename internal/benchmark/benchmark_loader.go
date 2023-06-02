package benchmark

import (
	"errors"
	"github.com/spf13/viper"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"strings"
)

const Path = "./contractsbenchmarks/benchmark/benchmarkconfigs"

type Benchmark struct {
	TargetSystem                       string
	ContactName                        string
	GossipNodeCount                    int32
	TotalNodeCount                     int32
	TotalClientCount                   int32
	TotalOrdererCount                  int32
	TotalSequencerCount                int32
	GossipIntervalMs                   int32
	TransactionTimeoutSecond           int32
	BlockTimeOutMs                     int32
	BlockTransactionSize               int32
	Profiling                          string
	ProfilingComponents                string
	Rounds                             []Round
	EndorsementPolicyOrgs              int32
	ProposalQueueConsumptionRateTPS    int32
	TransactionQueueConsumptionRateTPS int32
	QueueTickerDurationMS              int32
	ExtraEndorsementOrgs               int32
	OrgsPercentageIncreasedLoad        int32
	LoadIncreasePercentage             int32
}

type Round struct {
	ExperimentID                   string
	Label                          string
	BenchmarkFunctionName          string
	NumberOfClients                int
	NumberOfKeys                   int
	NumberOfKeysSecond             int
	NumberOfKeysThird              int
	TransactionsSendDurationSecond int
	TotalTransactions              int
	TotalSubmissionRate            int
	ReportImportance               bool
	CrdtObjectCount                string
	CrdtOperationPerObjectCount    string
	CrdtObjectType                 string
	Failures                       []Failure
}

type Failure struct {
	DurationS              int32
	FailureStartS          int32
	FailedOrgsCount        int32
	FailedClientPercentage int32
	FailureType            string
	NotifyClients          bool
	ClientStartAfterS      int32
}

func LoadBenchmark(name string) (*Benchmark, error) {
	viper.AddConfigPath(Path)
	viper.SetConfigName(name)
	viper.SetConfigType("yml")
	err := viper.ReadInConfig()
	if err != nil {
		return &Benchmark{}, err
	}
	benchmark := &Benchmark{}
	err = viper.Unmarshal(benchmark)
	if benchmark.EndorsementPolicyOrgs == 0 {
		logger.FatalLogger.Fatalln("Endorsement policy cannot be zero")
	}
	if benchmark.TotalNodeCount == 0 {
		logger.FatalLogger.Fatalln("Total node count cannot be zero")
	}
	if benchmark.TotalClientCount == 0 {
		logger.FatalLogger.Fatalln("Total client count cannot be zero")
	}
	if benchmark.GossipNodeCount >= benchmark.TotalNodeCount {
		logger.FatalLogger.Fatalln("Gossip node count cannot be bigger than total node count")
	}
	if benchmark.GossipNodeCount == 0 {
		logger.FatalLogger.Fatalln("Gossip node count cannot be zero")
	}
	if benchmark.GossipIntervalMs == 0 {
		logger.FatalLogger.Fatalln("Gossip interval ms cannot be zero")
	}
	if benchmark.ProposalQueueConsumptionRateTPS == 0 {
		logger.FatalLogger.Fatalln("Proposal queue size cannot be zero")
	}
	if benchmark.TransactionQueueConsumptionRateTPS == 0 {
		logger.FatalLogger.Fatalln("Transaction queue cannot be zero")
	}
	if benchmark.TransactionTimeoutSecond == 0 {
		logger.FatalLogger.Fatalln("Transaction timeout cannot be zero")
	}
	if benchmark.BlockTimeOutMs == 0 {
		logger.FatalLogger.Fatalln("Block time out cannot be zero")
	}
	if benchmark.BlockTransactionSize == 0 {
		logger.FatalLogger.Fatalln("Block size cannot be zero")
	}
	if benchmark.QueueTickerDurationMS == 0 {
		logger.FatalLogger.Fatalln("Queue ticker duration cannot be zero")
	}
	if benchmark.OrgsPercentageIncreasedLoad == 0 && benchmark.LoadIncreasePercentage > 0 {
		logger.FatalLogger.Fatalln("Orgs percentage increased load cannot be zero")
	}
	for _, round := range benchmark.Rounds {
		if round.TransactionsSendDurationSecond > 0 && round.TotalTransactions > 0 {
			logger.FatalLogger.Fatalln("BOTH TransactionsSendDurationSecond and TotalTransactions are set for ", round.Label)
		}
		if round.TransactionsSendDurationSecond == 0 && round.TotalTransactions == 0 {
			logger.FatalLogger.Fatalln("NONE of TransactionsSendDurationSecond or TotalTransactions are set for ", round.Label)
		}
		if round.NumberOfClients == 0 {
			logger.FatalLogger.Fatalln("NumberOfClients is not set for ", round.Label)
		}
		if round.NumberOfClients > int(benchmark.TotalClientCount) {
			logger.FatalLogger.Fatalln("Number of clients set  for this round is larger than total number of clients for ", round.Label)
		}
		if len(round.BenchmarkFunctionName) == 0 {
			logger.FatalLogger.Fatalln("BenchmarkFunctionName is not set for ", round.Label)
		}
		if round.TotalSubmissionRate == 0 {
			logger.FatalLogger.Fatalln("TotalSubmissionRate is not set for ", round.Label)
		}
		for _, failure := range round.Failures {
			if failure.FailureStartS == 0 {
				logger.FatalLogger.Fatalln("Failed start time second cannot be zero")
			}
			if failure.DurationS == 0 {
				logger.FatalLogger.Fatalln("Duration ms cannot be zero")
			}
			if failure.FailedOrgsCount == 0 && failure.FailedClientPercentage == 0 {
				logger.FatalLogger.Fatalln("Either failed orgs or failed client percentage must be set")
			}
			if len(failure.FailureType) == 0 {
				logger.FatalLogger.Fatalln("Failure type cannot be empty")
			}
		}
	}
	return benchmark, err
}

func GetTargetSystem(targetSystem string) (protos.TargetSystem, error) {
	targetSystem = strings.ToLower(targetSystem)
	switch targetSystem {
	case "orderlesschain":
		return protos.TargetSystem_ORDERLESSCHAIN, nil
	case "fabric":
		return protos.TargetSystem_FABRIC, nil
	case "fabriccrdt":
		return protos.TargetSystem_FABRICCRDT, nil
	case "bidl":
		return protos.TargetSystem_BIDL, nil
	case "synchotstuff":
		return protos.TargetSystem_SYNCHOTSTUFF, nil
	default:
		return 0, errors.New("the target system not found")
	}
}
