package config

import (
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gitlab.lrz.de/orderless/orderlesschain/internal/customcrypto/keygenerator"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/internal/profiling"
	"strings"
	"time"
)

var Config *Configuration

type Configuration struct {
	TargetSystem                         string `mapstructure:"TARGET_SYSTEM"`
	TransactionServerPort                string `mapstructure:"TRANSACTION_SERVER_PORT"`
	ProfilingServerPort                  string `mapstructure:"PROFILING_SERVER_PORT"`
	OrdererServerPort                    string `mapstructure:"ORDERER_SERVER_PORT"`
	SequencerServerPort                  string `mapstructure:"SEQUENCER_SERVER_PORT"`
	ProposalConnectionStreamPoolCount    int    `mapstructure:"PROPOSAL_CONNECTION_STREAM_POOL_COUNT"`
	TransactionConnectionStreamPoolCount int    `mapstructure:"TRANSACTION_CONNECTION_STREAM_POOL_COUNT"`
	TransactionConnectionPoolCount       int    `mapstructure:"TRANSACTION_CONNECTION_POOL_COUNT"`
	ProposalEventConnectionPoolCount     int    `mapstructure:"PROPOSAL_EVENT_CONNECTION_POOL_COUNT"`
	TransactionEventConnectionPoolCount  int    `mapstructure:"TRANSACTION_EVENT_CONNECTION_POOL_COUNT"`
	GossipConnectionPoolCount            int    `mapstructure:"GOSSIP_CONNECTION_POOL_COUNT"`
	OrdererConnectionPoolCount           int    `mapstructure:"ORDERER_CONNECTION_POOL_COUNT"`
	SequencerConnectionPoolCount         int    `mapstructure:"SEQUENCER_CONNECTION_POOL_COUNT"`
	ClientConnectionPoolCount            int    `mapstructure:"CLIENT_CONNECTION_POOL_COUNT"`
	InsideContractConnectionPoolCount    int    `mapstructure:"INSIDE_CONTRACT_CONNECTION_POOL_COUNT"`
	UUID                                 string `mapstructure:"UUID"`
	BlockTimeoutMS                       int    `mapstructure:"BLOCK_TIMEOUT_MS"`
	BlockTransactionSize                 int    `mapstructure:"BLOCK_TRANSACTION_SIZE"`
	IsRSAKeysCreated                     string `mapstructure:"IS_RSA_KEYS_CREATED"`
	GossipNodeCount                      int    `mapstructure:"GOSSIP_NODE_COUNT"`
	TotalOrdererCount                    int    `mapstructure:"TOTAL_ORDERER_COUNT"`
	TotalSequencerCount                  int    `mapstructure:"TOTAL_SEQUENCER_COUNT"`
	TotalNodeCount                       int    `mapstructure:"TOTAL_NODE_COUNT"`
	TotalClientCount                     int    `mapstructure:"TOTAL_CLIENT_COUNT"`
	TransactionTimeoutSecond             int    `mapstructure:"TRANSACTION_TIMEOUT_SECOND"`
	GossipIntervalMS                     int    `mapstructure:"GOSSIP_INTERVAL_MS"`
	ProposalQueueConsumptionRateTPS      int    `mapstructure:"PROPOSAL_QUEUE_CONSUMPTION_RATE_TPS"`
	TransactionQueueConsumptionRateTPS   int    `mapstructure:"TRANSACTION_QUEUE_CONSUMPTION_RATE_TPS"`
	QueueTickerDurationMS                int    `mapstructure:"QUEUE_TICKER_DURATION_MS"`
	EndorsementPolicy                    int    `mapstructure:"ENDORSEMENT_POLICY"`
	ExtraEndorsementOrgs                 int    `mapstructure:"EXTRA_ENDORSEMENT_ORGS"`
	ProfilingEnabled                     string `mapstructure:"PROFILING_ENABLED"`
	OrgsPercentageIncreasedLoad          int    `mapstructure:"ORGS_PERCENTAGE_INCREASED_LOAD"`
	LoadIncreasePercentage               int    `mapstructure:"LOAD_INCREASE_PERCENTAGE"`
	IPAddress                            string `mapstructure:"IP_ADDRESS"`
	IsTransactionProfilingEnabled        bool   `mapstructure:"TRANSACTION_PROFILING_ENABLED"`
	IsRealTimeCPUMemoryProfilingEnabled  bool   `mapstructure:"REAL_TIME_CPU_MEMORY_PROFILING_ENABLED"`
	TransactionTimeoutSecondConverted    time.Duration
	Nodes                                map[string]string
	Orderers                             map[string]string
	Sequencers                           map[string]string
	Clients                              map[string]string
	IsOrderlessChain                     bool
	IsFabric                             bool
	IsFabricCRDT                         bool
	IsBIDL                               bool
	IsSyncHotStuff                       bool
}

func init() {
	var err error
	Config, err = LoadConfig()
	if err != nil {
		logger.FatalLogger.Fatalln(err)
	}
}

func LoadConfig() (*Configuration, error) {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	err := viper.ReadInConfig()
	if err != nil {
		return &Configuration{}, err
	}
	appUUID := viper.GetString("UUID")
	if len(appUUID) == 0 {
		viper.Set("UUID", uuid.NewString())
		err = viper.WriteConfig()
		if err != nil {
			return &Configuration{}, err
		}
	}
	rsaPrivateKey := viper.GetString("IS_RSA_KEYS_CREATED")
	if len(rsaPrivateKey) == 0 {
		_, err = keygenerator.NewRSAKeys()
		if err != nil {
			return &Configuration{}, err
		}
		viper.Set("IS_RSA_KEYS_CREATED", "Yes")
		err = viper.WriteConfig()
		if err != nil {
			return &Configuration{}, err
		}
	}
	viper.SetConfigName("endpoints")
	viper.SetConfigType("yml")
	err = viper.MergeInConfig()
	if err != nil {
		return &Configuration{}, err
	}
	config := &Configuration{}
	err = viper.Unmarshal(config)
	if err != nil {
		return &Configuration{}, err
	}
	config.TargetSystem = strings.ToLower(config.TargetSystem)
	if config.TargetSystem == "orderlesschain" {
		config.IsOrderlessChain = true
	}
	if config.TargetSystem == "fabric" {
		config.IsFabric = true
	}
	if config.TargetSystem == "fabriccrdt" {
		config.IsFabricCRDT = true
	}
	if config.TargetSystem == "bidl" {
		config.IsBIDL = true
	}
	if config.TargetSystem == "synchotstuff" {
		config.IsSyncHotStuff = true
	}
	if config.ProfilingEnabled == "enable_server" {
		profiling.StartProfilingServer(config.ProfilingServerPort)
	} else if config.ProfilingEnabled == "cpu" {
		profiling.StartCPUProfiling()
	} else if config.ProfilingEnabled == "memory" {
		profiling.StartMemoryProfiling()
	} else if config.ProfilingEnabled == "bandwidth" {
		profiling.StartBandWithProfiling()
	}
	if config.TransactionTimeoutSecond > 0 {
		config.TransactionTimeoutSecondConverted = time.Duration(config.TransactionTimeoutSecond) * time.Second
	}
	if config.TotalNodeCount == 0 || config.TotalClientCount == 0 || config.TotalOrdererCount == 0 ||
		config.GossipNodeCount == 0 || config.EndorsementPolicy == 0 || config.QueueTickerDurationMS == 0 {
		logger.InfoLogger.Println("TotalNodeCount", config.TotalNodeCount)
		logger.InfoLogger.Println("TotalClientCount", config.TotalClientCount)
		logger.InfoLogger.Println("TotalOrdererCount", config.TotalOrdererCount)
		logger.InfoLogger.Println("TotalSequencerCount", config.TotalSequencerCount)
		logger.InfoLogger.Println("GossipNodeCount", config.GossipNodeCount)
		logger.InfoLogger.Println("EndorsementPolicy", config.EndorsementPolicy)
		logger.InfoLogger.Println("QueueTickerDurationMS", config.QueueTickerDurationMS)
		logger.FatalLogger.Fatalln("one of the above variables is not set in the config")
	}
	return config, nil
}
