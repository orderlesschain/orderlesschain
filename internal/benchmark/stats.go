package benchmark

import (
	"encoding/csv"
	"github.com/montanaflynn/stats"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type TransactionStat struct {
	transactionId          string
	startTime              time.Time
	endTime                time.Time
	latencyDuration        time.Duration
	latencyDurationFloat64 float64
	latencyMS              int64
	status                 string
	reportImportance       bool
}

type ClientStat struct {
	transactionsALL          map[string]*TransactionStat
	doneTransactionsALL      int
	succeededTransactionsALL int
	failedTransactionsALL    int
	submitRatePerSecALL      int
	submissionALL            map[string]int
	throughputPerSecALL      int
	throughputALL            map[string]int
	minLatencyMsALL          int64
	maxLatencyMsALL          int64
	averageLatencyMsALL      int64
	percentile01LatencyMsALL int64
	percentile05LatencyMsALL int64
	percentile10LatencyMsALL int64
	percentile25LatencyMsALL int64
	percentile50LatencyMsALL int64
	percentile75LatencyMsALL int64
	percentile90LatencyMsALL int64
	percentile95LatencyMsALL int64
	percentile99LatencyMsALL int64

	sentProposalBytesALL        int
	receivedProposalBytesALL    int
	sentTransactionBytesALL     int
	receivedTransactionBytesALL int
	TotalSentKBALL              int
	TotalReceivedKBALL          int
	TotalSentPerSecondKBALL     int
	TotalReceivedPerSecondKBALL int

	transactionsWrite          map[string]int
	doneTransactionsWrite      int
	succeededTransactionsWrite int
	failedTransactionsWrite    int
	submitRatePerSecWrite      int
	submissionWrite            map[string]int
	throughputPerSecWrite      int
	throughputWrite            map[string]int
	minLatencyMsWrite          int64
	maxLatencyMsWrite          int64
	averageLatencyMsWrite      int64
	percentile01LatencyMsWrite int64
	percentile05LatencyMsWrite int64
	percentile10LatencyMsWrite int64
	percentile25LatencyMsWrite int64
	percentile50LatencyMsWrite int64
	percentile75LatencyMsWrite int64
	percentile90LatencyMsWrite int64
	percentile95LatencyMsWrite int64
	percentile99LatencyMsWrite int64

	transactionsRead          map[string]int
	doneTransactionsRead      int
	succeededTransactionsRead int
	failedTransactionsRead    int
	submitRatePerSecRead      int
	submissionRead            map[string]int
	throughputPerSecRead      int
	throughputRead            map[string]int
	minLatencyMsRead          int64
	maxLatencyMsRead          int64
	averageLatencyMsRead      int64
	percentile01LatencyMsRead int64
	percentile05LatencyMsRead int64
	percentile10LatencyMsRead int64
	percentile25LatencyMsRead int64
	percentile50LatencyMsRead int64
	percentile75LatencyMsRead int64
	percentile90LatencyMsRead int64
	percentile95LatencyMsRead int64
	percentile99LatencyMsRead int64
}

type ExperimentSummary struct {
	clientStats map[string]*ClientStat

	submitRatePerSecALL      int
	throughputPerSecALL      int
	minLatencyMsALL          int64
	maxLatencyMsALL          int64
	averageLatencyMsALL      int64
	totalTransactionsALL     int
	succeededTransactionsALL int
	failedTransactionsALL    int
	percentile01LatencyMsALL int64
	percentile05LatencyMsALL int64
	percentile10LatencyMsALL int64
	percentile25LatencyMsALL int64
	percentile50LatencyMsALL int64
	percentile75LatencyMsALL int64
	percentile90LatencyMsALL int64
	percentile95LatencyMsALL int64
	percentile99LatencyMsALL int64

	sentProposalBytesALL        int
	receivedProposalBytesALL    int
	sentTransactionBytesALL     int
	receivedTransactionBytesALL int
	TotalSentKBALL              int
	TotalReceivedKBALL          int
	TotalSentPerSecondKBALL     int
	TotalReceivedPerSecondKBALL int

	submitRatePerSecWrite      int
	throughputPerSecWrite      int
	minLatencyMsWrite          int64
	maxLatencyMsWrite          int64
	averageLatencyMsWrite      int64
	totalTransactionsWrite     int
	succeededTransactionsWrite int
	failedTransactionsWrite    int
	percentile01LatencyMsWrite int64
	percentile05LatencyMsWrite int64
	percentile10LatencyMsWrite int64
	percentile25LatencyMsWrite int64
	percentile50LatencyMsWrite int64
	percentile75LatencyMsWrite int64
	percentile90LatencyMsWrite int64
	percentile95LatencyMsWrite int64
	percentile99LatencyMsWrite int64

	submitRatePerSecRead      int
	throughputPerSecRead      int
	minLatencyMsRead          int64
	maxLatencyMsRead          int64
	averageLatencyMsRead      int64
	totalTransactionsRead     int
	succeededTransactionsRead int
	failedTransactionsRead    int
	percentile01LatencyMsRead int64
	percentile05LatencyMsRead int64
	percentile10LatencyMsRead int64
	percentile25LatencyMsRead int64
	percentile50LatencyMsRead int64
	percentile75LatencyMsRead int64
	percentile90LatencyMsRead int64
	percentile95LatencyMsRead int64
	percentile99LatencyMsRead int64
}

func CalculateExperimentsStats(reportPath string) *ExperimentSummary {
	summary := &ExperimentSummary{
		clientStats:       map[string]*ClientStat{},
		minLatencyMsALL:   math.MaxInt64,
		minLatencyMsWrite: math.MaxInt64,
		minLatencyMsRead:  math.MaxInt64,
	}
	var summaryLatencyDurationsFloat64ALL []float64
	var summaryLatencyDurationsFloat64Write []float64
	var summaryLatencyDurationsFloat64Read []float64

	err := filepath.Walk(reportPath, func(path string, reportInfo os.FileInfo, err error) error {
		if err != nil {
			logger.ErrorLogger.Println(err)
			return nil
		}
		reportFile, err := os.Open(filepath.Join(reportPath, reportInfo.Name()))
		if err != nil {
			return nil
		}
		reportLines, err := csv.NewReader(reportFile).ReadAll()
		if err != nil {
			logger.ErrorLogger.Println("Couldn't read the csv file", err)
			return nil
		}
		tempClientStat := &ClientStat{
			transactionsALL: map[string]*TransactionStat{},
			submissionALL:   map[string]int{},
			throughputALL:   map[string]int{},
			minLatencyMsALL: math.MaxInt64,

			transactionsWrite: map[string]int{},
			submissionWrite:   map[string]int{},
			throughputWrite:   map[string]int{},
			minLatencyMsWrite: math.MaxInt64,

			transactionsRead: map[string]int{},
			submissionRead:   map[string]int{},
			throughputRead:   map[string]int{},
			minLatencyMsRead: math.MaxInt64,
		}
		var latencyDurationsFloat64ALL []float64
		var latencyDurationsFloat64Write []float64
		var latencyDurationsFloat64Read []float64

		for _, line := range reportLines {
			transactionId := line[0]
			startTimeUnixNano, _ := strconv.ParseInt(line[1], 10, 64)
			startTime := time.Unix(0, startTimeUnixNano)
			endTimeUnixNano, _ := strconv.ParseInt(line[2], 10, 64)
			endTime := time.Unix(0, endTimeUnixNano)
			latencyDurations, _ := time.ParseDuration(line[3])
			latencyDurationFloat64 := float64(latencyDurations)
			latencyMS := int64(latencyDurations / time.Millisecond)
			status := line[4]
			reportImportance, _ := strconv.ParseBool(line[5])
			writeReadType := line[6]
			sentProposalBytes, _ := strconv.Atoi(line[7])
			receivedProposalBytes, _ := strconv.Atoi(line[8])
			sentTransactionBytes, _ := strconv.Atoi(line[9])
			receivedTransactionBytes, _ := strconv.Atoi(line[10])
			if !reportImportance {
				continue
			}
			writeTransaction := false
			readTransaction := false
			if len(writeReadType) > 0 {
				writeReadTypeInt := protos.ProposalRequest_WriteReadTransaction_value[line[6]]
				if writeReadTypeInt == int32(protos.ProposalRequest_Write) {
					writeTransaction = true
				}
				if writeReadTypeInt == int32(protos.ProposalRequest_Read) {
					readTransaction = true
				}
			}
			startTimeString := startTime.Format("20060102150405")
			endTimeString := endTime.Format("20060102150405")

			if _, ok := tempClientStat.submissionALL[startTimeString]; ok {
				tempClientStat.submissionALL[startTimeString]++
			} else {
				tempClientStat.submissionALL[startTimeString] = 1
			}
			if writeTransaction {
				if _, ok := tempClientStat.submissionWrite[startTimeString]; ok {
					tempClientStat.submissionWrite[startTimeString]++
				} else {
					tempClientStat.submissionWrite[startTimeString] = 1
				}
			}
			if readTransaction {
				if _, ok := tempClientStat.submissionRead[startTimeString]; ok {
					tempClientStat.submissionRead[startTimeString]++
				} else {
					tempClientStat.submissionRead[startTimeString] = 1
				}
			}
			if status == "SUCCEEDED" {
				tempClientStat.doneTransactionsALL++
				tempClientStat.succeededTransactionsALL++
				if writeTransaction {
					tempClientStat.doneTransactionsWrite++
					tempClientStat.succeededTransactionsWrite++
				}
				if readTransaction {
					tempClientStat.doneTransactionsRead++
					tempClientStat.succeededTransactionsRead++
				}
			} else if status == "FAILED_TIMEOUT" {
				tempClientStat.failedTransactionsALL++
				if writeTransaction {
					tempClientStat.failedTransactionsWrite++
				}
				if readTransaction {
					tempClientStat.failedTransactionsRead++
				}
				continue
			} else {
				tempClientStat.failedTransactionsALL++
				if writeTransaction {
					tempClientStat.failedTransactionsWrite++
				}
				if readTransaction {
					tempClientStat.failedTransactionsRead++
				}
				continue
			}
			tempClientStat.transactionsALL[transactionId] = &TransactionStat{
				transactionId:          transactionId,
				startTime:              startTime,
				endTime:                endTime,
				latencyDuration:        latencyDurations,
				latencyDurationFloat64: latencyDurationFloat64,
				latencyMS:              latencyMS,
				status:                 status,
			}
			if writeTransaction {
				tempClientStat.transactionsWrite[transactionId] = 1
			}
			if readTransaction {
				tempClientStat.transactionsRead[transactionId] = 1
			}

			tempClientStat.sentProposalBytesALL += sentProposalBytes
			tempClientStat.receivedProposalBytesALL += receivedProposalBytes
			tempClientStat.sentTransactionBytesALL += sentTransactionBytes
			tempClientStat.receivedTransactionBytesALL += receivedTransactionBytes

			if _, ok := tempClientStat.throughputALL[endTimeString]; ok {
				tempClientStat.throughputALL[endTimeString]++
			} else {
				tempClientStat.throughputALL[endTimeString] = 1
			}
			if tempClientStat.maxLatencyMsALL < latencyMS {
				tempClientStat.maxLatencyMsALL = latencyMS
			}
			if tempClientStat.minLatencyMsALL > latencyMS {
				tempClientStat.minLatencyMsALL = latencyMS
			}
			tempClientStat.averageLatencyMsALL += latencyMS

			latencyDurationsFloat64ALL = append(latencyDurationsFloat64ALL, latencyDurationFloat64)
			summaryLatencyDurationsFloat64ALL = append(summaryLatencyDurationsFloat64ALL, latencyDurationFloat64)

			if writeTransaction {
				if _, ok := tempClientStat.throughputWrite[endTimeString]; ok {
					tempClientStat.throughputWrite[endTimeString]++
				} else {
					tempClientStat.throughputWrite[endTimeString] = 1
				}
				if tempClientStat.maxLatencyMsWrite < latencyMS {
					tempClientStat.maxLatencyMsWrite = latencyMS
				}
				if tempClientStat.minLatencyMsWrite > latencyMS {
					tempClientStat.minLatencyMsWrite = latencyMS
				}
				tempClientStat.averageLatencyMsWrite += latencyMS

				latencyDurationsFloat64Write = append(latencyDurationsFloat64Write, latencyDurationFloat64)
				summaryLatencyDurationsFloat64Write = append(summaryLatencyDurationsFloat64Write, latencyDurationFloat64)
			}

			if readTransaction {
				if _, ok := tempClientStat.throughputRead[endTimeString]; ok {
					tempClientStat.throughputRead[endTimeString]++
				} else {
					tempClientStat.throughputRead[endTimeString] = 1
				}
				if tempClientStat.maxLatencyMsRead < latencyMS {
					tempClientStat.maxLatencyMsRead = latencyMS
				}
				if tempClientStat.minLatencyMsRead > latencyMS {
					tempClientStat.minLatencyMsRead = latencyMS
				}
				tempClientStat.averageLatencyMsRead += latencyMS

				latencyDurationsFloat64Read = append(latencyDurationsFloat64Read, latencyDurationFloat64)
				summaryLatencyDurationsFloat64Read = append(summaryLatencyDurationsFloat64Read, latencyDurationFloat64)
			}
		}
		if len(tempClientStat.transactionsALL) == 0 {
			return nil
		}
		if tempPercentile99, err := stats.Percentile(latencyDurationsFloat64ALL, 99); err != nil {
			logger.ErrorLogger.Println(err)
			return nil
		} else {
			tempClientStat.percentile99LatencyMsALL = int64(tempPercentile99 / float64(time.Millisecond))

			tempPercentile01, _ := stats.Percentile(latencyDurationsFloat64ALL, 1)
			tempClientStat.percentile01LatencyMsALL = int64(tempPercentile01 / float64(time.Millisecond))

			tempPercentile05, _ := stats.Percentile(latencyDurationsFloat64ALL, 5)
			tempClientStat.percentile05LatencyMsALL = int64(tempPercentile05 / float64(time.Millisecond))

			tempPercentile10, _ := stats.Percentile(latencyDurationsFloat64ALL, 10)
			tempClientStat.percentile10LatencyMsALL = int64(tempPercentile10 / float64(time.Millisecond))

			tempPercentile25, _ := stats.Percentile(latencyDurationsFloat64ALL, 25)
			tempClientStat.percentile25LatencyMsALL = int64(tempPercentile25 / float64(time.Millisecond))

			tempPercentile50, _ := stats.Percentile(latencyDurationsFloat64ALL, 50)
			tempClientStat.percentile50LatencyMsALL = int64(tempPercentile50 / float64(time.Millisecond))

			tempPercentile75, _ := stats.Percentile(latencyDurationsFloat64ALL, 75)
			tempClientStat.percentile75LatencyMsALL = int64(tempPercentile75 / float64(time.Millisecond))

			tempPercentile90, _ := stats.Percentile(latencyDurationsFloat64ALL, 90)
			tempClientStat.percentile90LatencyMsALL = int64(tempPercentile90 / float64(time.Millisecond))

			tempPercentile95, _ := stats.Percentile(latencyDurationsFloat64ALL, 95)
			tempClientStat.percentile95LatencyMsALL = int64(tempPercentile95 / float64(time.Millisecond))
		}
		for _, submission := range tempClientStat.submissionALL {
			tempClientStat.submitRatePerSecALL += submission
		}
		if len(tempClientStat.submissionALL) != 0 && len(tempClientStat.submissionALL) > 1 {
			tempClientStat.submitRatePerSecALL = tempClientStat.submitRatePerSecALL / (len(tempClientStat.submissionALL) - 1)
		}
		for _, throughput := range tempClientStat.throughputALL {
			tempClientStat.throughputPerSecALL += throughput
		}
		if len(tempClientStat.throughputALL) != 0 && len(tempClientStat.throughputALL) > 1 {
			tempClientStat.throughputPerSecALL = tempClientStat.throughputPerSecALL / (len(tempClientStat.throughputALL) - 1)
		}
		if tempClientStat.doneTransactionsALL != 0 {
			tempClientStat.averageLatencyMsALL = tempClientStat.averageLatencyMsALL / int64(tempClientStat.doneTransactionsALL)
		}

		if len(tempClientStat.submissionALL) != 0 && len(tempClientStat.submissionALL) > 1 {
			tempClientStat.TotalSentKBALL = (tempClientStat.sentProposalBytesALL + tempClientStat.sentTransactionBytesALL) / 1024
			tempClientStat.TotalSentPerSecondKBALL = tempClientStat.TotalSentKBALL / (len(tempClientStat.submissionALL) - 1)
		}
		if len(tempClientStat.throughputALL) != 0 && len(tempClientStat.throughputALL) > 1 {
			tempClientStat.TotalReceivedKBALL = (tempClientStat.receivedProposalBytesALL + tempClientStat.receivedTransactionBytesALL) / 1024
			tempClientStat.TotalReceivedPerSecondKBALL = tempClientStat.TotalReceivedKBALL / (len(tempClientStat.throughputALL) - 1)
		}

		summary.clientStats[reportInfo.Name()] = tempClientStat

		summary.sentProposalBytesALL += tempClientStat.sentProposalBytesALL
		summary.receivedProposalBytesALL += tempClientStat.receivedProposalBytesALL
		summary.sentTransactionBytesALL += tempClientStat.sentTransactionBytesALL
		summary.receivedTransactionBytesALL += tempClientStat.receivedTransactionBytesALL

		summary.throughputPerSecALL += tempClientStat.throughputPerSecALL
		summary.submitRatePerSecALL += tempClientStat.submitRatePerSecALL
		summary.averageLatencyMsALL += tempClientStat.averageLatencyMsALL
		if summary.maxLatencyMsALL < tempClientStat.maxLatencyMsALL {
			summary.maxLatencyMsALL = tempClientStat.maxLatencyMsALL
		}
		if summary.minLatencyMsALL > tempClientStat.minLatencyMsALL {
			summary.minLatencyMsALL = tempClientStat.minLatencyMsALL
		}
		summary.totalTransactionsALL += len(tempClientStat.transactionsALL)
		summary.failedTransactionsALL += tempClientStat.failedTransactionsALL
		summary.succeededTransactionsALL += tempClientStat.succeededTransactionsALL
		if len(tempClientStat.transactionsWrite) != 0 {
			if tempPercentile, err := stats.Percentile(latencyDurationsFloat64Write, 99); err != nil {
				logger.ErrorLogger.Println(err)
				return nil
			} else {
				tempClientStat.percentile99LatencyMsWrite = int64(tempPercentile / float64(time.Millisecond))

				tempPercentile01, _ := stats.Percentile(latencyDurationsFloat64Write, 1)
				tempClientStat.percentile01LatencyMsWrite = int64(tempPercentile01 / float64(time.Millisecond))

				tempPercentile05, _ := stats.Percentile(latencyDurationsFloat64Write, 5)
				tempClientStat.percentile05LatencyMsWrite = int64(tempPercentile05 / float64(time.Millisecond))

				tempPercentile10, _ := stats.Percentile(latencyDurationsFloat64Write, 10)
				tempClientStat.percentile10LatencyMsWrite = int64(tempPercentile10 / float64(time.Millisecond))

				tempPercentile25, _ := stats.Percentile(latencyDurationsFloat64Write, 25)
				tempClientStat.percentile25LatencyMsWrite = int64(tempPercentile25 / float64(time.Millisecond))

				tempPercentile50, _ := stats.Percentile(latencyDurationsFloat64Write, 50)
				tempClientStat.percentile50LatencyMsWrite = int64(tempPercentile50 / float64(time.Millisecond))

				tempPercentile75, _ := stats.Percentile(latencyDurationsFloat64Write, 75)
				tempClientStat.percentile75LatencyMsWrite = int64(tempPercentile75 / float64(time.Millisecond))

				tempPercentile90, _ := stats.Percentile(latencyDurationsFloat64Write, 90)
				tempClientStat.percentile90LatencyMsWrite = int64(tempPercentile90 / float64(time.Millisecond))

				tempPercentile95, _ := stats.Percentile(latencyDurationsFloat64Write, 95)
				tempClientStat.percentile95LatencyMsWrite = int64(tempPercentile95 / float64(time.Millisecond))
			}
			for _, submission := range tempClientStat.submissionWrite {
				tempClientStat.submitRatePerSecWrite += submission
			}
			if len(tempClientStat.submissionWrite) != 0 && len(tempClientStat.submissionWrite) > 1 {
				tempClientStat.submitRatePerSecWrite = tempClientStat.submitRatePerSecWrite / (len(tempClientStat.submissionWrite) - 1)
			}
			for _, throughput := range tempClientStat.throughputWrite {
				tempClientStat.throughputPerSecWrite += throughput
			}

			if len(tempClientStat.throughputWrite) != 0 && len(tempClientStat.throughputWrite) > 1 {
				tempClientStat.throughputPerSecWrite = tempClientStat.throughputPerSecWrite / (len(tempClientStat.throughputWrite) - 1)
			}
			if tempClientStat.doneTransactionsWrite != 0 {
				tempClientStat.averageLatencyMsWrite = tempClientStat.averageLatencyMsWrite / int64(tempClientStat.doneTransactionsWrite)
			}
			summary.clientStats[reportInfo.Name()] = tempClientStat

			summary.throughputPerSecWrite += tempClientStat.throughputPerSecWrite
			summary.submitRatePerSecWrite += tempClientStat.submitRatePerSecWrite
			summary.averageLatencyMsWrite += tempClientStat.averageLatencyMsWrite
			if summary.maxLatencyMsWrite < tempClientStat.maxLatencyMsWrite {
				summary.maxLatencyMsWrite = tempClientStat.maxLatencyMsWrite
			}
			if summary.minLatencyMsWrite > tempClientStat.minLatencyMsWrite {
				summary.minLatencyMsWrite = tempClientStat.minLatencyMsWrite
			}
			summary.totalTransactionsWrite += len(tempClientStat.transactionsWrite)
			summary.failedTransactionsWrite += tempClientStat.failedTransactionsWrite
			summary.succeededTransactionsWrite += tempClientStat.succeededTransactionsWrite
		}
		if len(tempClientStat.transactionsRead) != 0 {
			if tempPercentile, err := stats.Percentile(latencyDurationsFloat64Read, 99); err != nil {
				logger.ErrorLogger.Println(err)
				return nil
			} else {
				tempClientStat.percentile99LatencyMsRead = int64(tempPercentile / float64(time.Millisecond))

				tempPercentile01, _ := stats.Percentile(latencyDurationsFloat64Read, 1)
				tempClientStat.percentile01LatencyMsRead = int64(tempPercentile01 / float64(time.Millisecond))

				tempPercentile05, _ := stats.Percentile(latencyDurationsFloat64Read, 5)
				tempClientStat.percentile05LatencyMsRead = int64(tempPercentile05 / float64(time.Millisecond))

				tempPercentile10, _ := stats.Percentile(latencyDurationsFloat64Read, 10)
				tempClientStat.percentile10LatencyMsRead = int64(tempPercentile10 / float64(time.Millisecond))

				tempPercentile25, _ := stats.Percentile(latencyDurationsFloat64Read, 25)
				tempClientStat.percentile25LatencyMsRead = int64(tempPercentile25 / float64(time.Millisecond))

				tempPercentile50, _ := stats.Percentile(latencyDurationsFloat64Read, 50)
				tempClientStat.percentile50LatencyMsRead = int64(tempPercentile50 / float64(time.Millisecond))

				tempPercentile75, _ := stats.Percentile(latencyDurationsFloat64Read, 75)
				tempClientStat.percentile75LatencyMsRead = int64(tempPercentile75 / float64(time.Millisecond))

				tempPercentile90, _ := stats.Percentile(latencyDurationsFloat64Read, 90)
				tempClientStat.percentile90LatencyMsRead = int64(tempPercentile90 / float64(time.Millisecond))

				tempPercentile95, _ := stats.Percentile(latencyDurationsFloat64Read, 95)
				tempClientStat.percentile95LatencyMsRead = int64(tempPercentile95 / float64(time.Millisecond))
			}
			for _, submission := range tempClientStat.submissionRead {
				tempClientStat.submitRatePerSecRead += submission
			}
			if len(tempClientStat.submissionRead) != 0 && len(tempClientStat.submissionRead) > 1 {
				tempClientStat.submitRatePerSecRead = tempClientStat.submitRatePerSecRead / (len(tempClientStat.submissionRead) - 1)
			}
			for _, throughput := range tempClientStat.throughputRead {
				tempClientStat.throughputPerSecRead += throughput
			}

			if len(tempClientStat.throughputRead) != 0 && len(tempClientStat.throughputRead) > 1 {
				tempClientStat.throughputPerSecRead = tempClientStat.throughputPerSecRead / (len(tempClientStat.throughputRead) - 1)
			}
			if tempClientStat.doneTransactionsRead != 0 {
				tempClientStat.averageLatencyMsRead = tempClientStat.averageLatencyMsRead / int64(tempClientStat.doneTransactionsRead)
			}
			summary.clientStats[reportInfo.Name()] = tempClientStat

			summary.throughputPerSecRead += tempClientStat.throughputPerSecRead
			summary.submitRatePerSecRead += tempClientStat.submitRatePerSecRead
			summary.averageLatencyMsRead += tempClientStat.averageLatencyMsRead
			if summary.maxLatencyMsRead < tempClientStat.maxLatencyMsRead {
				summary.maxLatencyMsRead = tempClientStat.maxLatencyMsRead
			}
			if summary.minLatencyMsRead > tempClientStat.minLatencyMsRead {
				summary.minLatencyMsRead = tempClientStat.minLatencyMsRead
			}
			summary.totalTransactionsRead += len(tempClientStat.transactionsRead)
			summary.failedTransactionsRead += tempClientStat.failedTransactionsRead
			summary.succeededTransactionsRead += tempClientStat.succeededTransactionsRead
		}
		return nil
	})
	if err != nil {
		logger.ErrorLogger.Println(err)
	}

	clientStatsLength := int64(len(summary.clientStats))
	if clientStatsLength != 0 {
		summary.averageLatencyMsALL = summary.averageLatencyMsALL / clientStatsLength
		summary.averageLatencyMsWrite = summary.averageLatencyMsWrite / clientStatsLength
		summary.averageLatencyMsRead = summary.averageLatencyMsRead / clientStatsLength

		summary.TotalSentKBALL = (summary.sentProposalBytesALL + summary.sentTransactionBytesALL) / 1024
		summary.TotalSentPerSecondKBALL = summary.TotalSentKBALL / int(clientStatsLength)

		summary.TotalReceivedKBALL = (summary.receivedProposalBytesALL + summary.receivedTransactionBytesALL) / 1024
		summary.TotalReceivedPerSecondKBALL = summary.TotalReceivedKBALL / int(clientStatsLength)
	}

	if tempPercentile, err := stats.Percentile(summaryLatencyDurationsFloat64ALL, 99); err != nil {
		logger.ErrorLogger.Println(err)
	} else {
		summary.percentile99LatencyMsALL = int64(tempPercentile / float64(time.Millisecond))

		tempPercentile01, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 1)
		summary.percentile01LatencyMsALL = int64(tempPercentile01 / float64(time.Millisecond))

		tempPercentile05, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 5)
		summary.percentile05LatencyMsALL = int64(tempPercentile05 / float64(time.Millisecond))

		tempPercentile10, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 10)
		summary.percentile10LatencyMsALL = int64(tempPercentile10 / float64(time.Millisecond))

		tempPercentile25, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 25)
		summary.percentile25LatencyMsALL = int64(tempPercentile25 / float64(time.Millisecond))

		tempPercentile50, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 50)
		summary.percentile50LatencyMsALL = int64(tempPercentile50 / float64(time.Millisecond))

		tempPercentile75, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 75)
		summary.percentile75LatencyMsALL = int64(tempPercentile75 / float64(time.Millisecond))

		tempPercentile90, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 90)
		summary.percentile90LatencyMsALL = int64(tempPercentile90 / float64(time.Millisecond))

		tempPercentile95, _ := stats.Percentile(summaryLatencyDurationsFloat64ALL, 95)
		summary.percentile95LatencyMsALL = int64(tempPercentile95 / float64(time.Millisecond))
	}
	if tempPercentile, err := stats.Percentile(summaryLatencyDurationsFloat64Write, 99); err != nil {
		logger.ErrorLogger.Println(err)
	} else {
		summary.percentile99LatencyMsWrite = int64(tempPercentile / float64(time.Millisecond))

		tempPercentile01, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 1)
		summary.percentile01LatencyMsWrite = int64(tempPercentile01 / float64(time.Millisecond))

		tempPercentile05, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 5)
		summary.percentile05LatencyMsWrite = int64(tempPercentile05 / float64(time.Millisecond))

		tempPercentile10, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 10)
		summary.percentile10LatencyMsWrite = int64(tempPercentile10 / float64(time.Millisecond))

		tempPercentile25, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 25)
		summary.percentile25LatencyMsWrite = int64(tempPercentile25 / float64(time.Millisecond))

		tempPercentile50, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 50)
		summary.percentile50LatencyMsWrite = int64(tempPercentile50 / float64(time.Millisecond))

		tempPercentile75, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 75)
		summary.percentile75LatencyMsWrite = int64(tempPercentile75 / float64(time.Millisecond))

		tempPercentile90, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 90)
		summary.percentile90LatencyMsWrite = int64(tempPercentile90 / float64(time.Millisecond))

		tempPercentile95, _ := stats.Percentile(summaryLatencyDurationsFloat64Write, 95)
		summary.percentile95LatencyMsWrite = int64(tempPercentile95 / float64(time.Millisecond))
	}
	if tempPercentile, err := stats.Percentile(summaryLatencyDurationsFloat64Read, 99); err != nil {
		logger.ErrorLogger.Println(err)
	} else {
		summary.percentile99LatencyMsRead = int64(tempPercentile / float64(time.Millisecond))

		tempPercentile01, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 1)
		summary.percentile01LatencyMsRead = int64(tempPercentile01 / float64(time.Millisecond))

		tempPercentile05, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 5)
		summary.percentile05LatencyMsRead = int64(tempPercentile05 / float64(time.Millisecond))

		tempPercentile10, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 10)
		summary.percentile10LatencyMsRead = int64(tempPercentile10 / float64(time.Millisecond))

		tempPercentile25, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 25)
		summary.percentile25LatencyMsRead = int64(tempPercentile25 / float64(time.Millisecond))

		tempPercentile50, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 50)
		summary.percentile50LatencyMsRead = int64(tempPercentile50 / float64(time.Millisecond))

		tempPercentile75, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 75)
		summary.percentile75LatencyMsRead = int64(tempPercentile75 / float64(time.Millisecond))

		tempPercentile90, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 90)
		summary.percentile90LatencyMsRead = int64(tempPercentile90 / float64(time.Millisecond))

		tempPercentile95, _ := stats.Percentile(summaryLatencyDurationsFloat64Read, 95)
		summary.percentile95LatencyMsRead = int64(tempPercentile95 / float64(time.Millisecond))
	}

	return summary
}
