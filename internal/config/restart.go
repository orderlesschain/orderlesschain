package config

import (
	"github.com/spf13/viper"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"os"
	"path"
	"path/filepath"
	"time"
)

func UpdateModeAndRestart(mode *protos.OperationMode) {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	err := viper.ReadInConfig()
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	switch mode.TargetSystem {
	case protos.TargetSystem_ORDERLESSCHAIN:
		viper.Set("TARGET_SYSTEM", "orderlesschain")
	case protos.TargetSystem_FABRIC:
		viper.Set("TARGET_SYSTEM", "fabric")
	case protos.TargetSystem_FABRICCRDT:
		viper.Set("TARGET_SYSTEM", "fabriccrdt")
	case protos.TargetSystem_BIDL:
		viper.Set("TARGET_SYSTEM", "bidl")
	case protos.TargetSystem_SYNCHOTSTUFF:
		viper.Set("TARGET_SYSTEM", "synchotstuff")
	}
	if mode.GossipNodeCount > 0 {
		viper.Set("GOSSIP_NODE_COUNT", mode.GossipNodeCount)
	}
	if mode.EndorsementPolicy > 0 {
		viper.Set("ENDORSEMENT_POLICY", mode.EndorsementPolicy)
	}
	if mode.TotalOrdererCount > 0 {
		viper.Set("TOTAL_ORDERER_COUNT", mode.TotalOrdererCount)
	}
	if mode.TotalSequencerCount > 0 {
		viper.Set("TOTAL_SEQUENCER_COUNT", mode.TotalSequencerCount)
	}
	if mode.TotalNodeCount > 0 {
		viper.Set("TOTAL_NODE_COUNT", mode.TotalNodeCount)
	}
	if mode.TotalClientCount > 0 {
		viper.Set("TOTAL_CLIENT_COUNT", mode.TotalClientCount)
	}
	if mode.GossipIntervalMs > 0 {
		viper.Set("GOSSIP_INTERVAL_MS", mode.GossipIntervalMs)
	}
	if mode.TransactionTimeoutSecond > 0 {
		viper.Set("TRANSACTION_TIMEOUT_SECOND", mode.TransactionTimeoutSecond)
	}
	if mode.BlockTimeOutMs > 0 {
		viper.Set("BLOCK_TIMEOUT_MS", mode.BlockTimeOutMs)
	}
	if mode.BlockTransactionSize > 0 {
		viper.Set("BLOCK_TRANSACTION_SIZE", mode.BlockTransactionSize)
	}
	if mode.ProposalQueueConsumptionRateTps > 0 {
		viper.Set("PROPOSAL_QUEUE_CONSUMPTION_RATE_TPS", mode.ProposalQueueConsumptionRateTps)
	}
	if mode.TransactionQueueConsumptionRateTps > 0 {
		viper.Set("TRANSACTION_QUEUE_CONSUMPTION_RATE_TPS", mode.TransactionQueueConsumptionRateTps)
	}
	if mode.QueueTickerDurationMs > 0 {
		viper.Set("QUEUE_TICKER_DURATION_MS", mode.QueueTickerDurationMs)
	}
	viper.Set("EXTRA_ENDORSEMENT_ORGS", mode.ExtraEndorsementOrgs)
	if len(mode.ProfilingEnabled) > 0 {
		viper.Set("PROFILING_ENABLED", mode.ProfilingEnabled)
	} else {
		viper.Set("PROFILING_ENABLED", "not_enabled")
	}
	viper.Set("ORGS_PERCENTAGE_INCREASED_LOAD", mode.OrgsPercentageIncreasedLoad)
	viper.Set("LOAD_INCREASE_PERCENTAGE", mode.LoadIncreasePercentage)
	err = viper.WriteConfig()
	if err != nil {
		logger.ErrorLogger.Println(err)
		return
	}
	logger.InfoLogger.Println("Data Deleted - Restarting for", viper.Get("TARGET_SYSTEM"),
		"Benchmark:", mode.Benchmark,
		"Total Node Count:", viper.Get("TOTAL_NODE_COUNT"),
		"Total Client Count:", viper.Get("TOTAL_CLIENT_COUNT"),
		"Gossip Node Count:", viper.Get("GOSSIP_NODE_COUNT"),
		"Gossip Interval MS:", viper.Get("GOSSIP_INTERVAL_MS"),
		"Transaction Timeout Second:", viper.Get("TRANSACTION_TIMEOUT_SECOND"),
		"Block Timeout MS:", viper.Get("BLOCK_TIMEOUT_MS"),
		"Block Transaction Size:", viper.Get("BLOCK_TRANSACTION_SIZE"),
		"Queue Ticker Duration MS:", viper.Get("QUEUE_TICKER_DURATION_MS"),
		"Proposal Queue Consumption Rate TPS:", viper.Get("PROPOSAL_QUEUE_CONSUMPTION_RATE_TPS"),
		"Transaction Queue Consumption Rate TPS:", viper.Get("TRANSACTION_QUEUE_CONSUMPTION_RATE_TPS"))
	time.Sleep(1 * time.Second)
	os.Exit(0)
}

func RemoveSavedData() error {
	dataPath := filepath.Join("./data/")
	dirRead, err := os.Open(dataPath)
	if err != nil {
		return err
	}
	dirFiles, err := dirRead.Readdir(0)
	if err != nil {
		return err
	}
	for index := range dirFiles {
		if err = os.RemoveAll(path.Join(dataPath, dirFiles[index].Name())); err != nil {
			logger.ErrorLogger.Println(err)
		}
	}
	if err = os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return err
	}
	return nil
}
