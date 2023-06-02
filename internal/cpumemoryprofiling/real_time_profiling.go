package cpumemoryprofiling

import (
	"bufio"
	"gitlab.lrz.de/orderless/orderlesschain/internal/config"
	"gitlab.lrz.de/orderless/orderlesschain/internal/logger"
	protos "gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const monitoringIntervalSecond = 1

type CPUUsage struct {
	PrevIdleTime       uint64
	PrevTotalTime      uint64
	CPUUsagePercentage int64
}

type MemoryUsage struct {
	AllocatedHeap  int64
	HeapInUse      int64
	TotalAllocated int64
	SystemMemory   int64
}

type CPUMemoryProfiler struct {
	memoryRecordLock *sync.Mutex
	CPUMemoryRecords map[string]*protos.MemoryCPUProfile
}

func InitCPUMemoryProfiler() *CPUMemoryProfiler {
	tempProfiler := &CPUMemoryProfiler{
		memoryRecordLock: &sync.Mutex{},
		CPUMemoryRecords: map[string]*protos.MemoryCPUProfile{},
	}
	tempProfiler.startMonitoringCPUMemory()
	return tempProfiler
}

func (m *CPUMemoryProfiler) startMonitoringCPUMemory() {
	if config.Config.IsRealTimeCPUMemoryProfilingEnabled {
		logger.InfoLogger.Println("cpu memory profiling is running")
		monitoringInterval := monitoringIntervalSecond * time.Second
		cpuUsage := &CPUUsage{}
		go func(cpuUsage *CPUUsage) {
			for {
				var err error
				cpuUsage, err = m.monitorCPUUsage(cpuUsage)
				if err != nil {
					return
				}
				memoryUsage, err := m.monitorMemoryUsage()
				if err != nil {
					return
				}
				tempRecord := &protos.MemoryCPUProfile{
					Timestamp:          time.Now().Format("20060102150405"),
					CpuUsagePercentage: cpuUsage.CPUUsagePercentage,
					AllocatedHeap:      memoryUsage.AllocatedHeap,
					HeapInUse:          memoryUsage.HeapInUse,
					TotalAllocated:     memoryUsage.TotalAllocated,
					SystemMemory:       memoryUsage.SystemMemory,
				}
				m.memoryRecordLock.Lock()
				m.CPUMemoryRecords[tempRecord.Timestamp] = tempRecord
				m.memoryRecordLock.Unlock()
				time.Sleep(monitoringInterval)
			}
		}(cpuUsage)
	}
}

func (m *CPUMemoryProfiler) monitorCPUUsage(cpuUsage *CPUUsage) (*CPUUsage, error) {
	procStat, err := os.Open("/proc/stat")
	if err != nil {
		logger.ErrorLogger.Println(err)
		return nil, err
	}
	procStatScanner := bufio.NewScanner(procStat)
	procStatScanner.Scan()
	cpuInfo := procStatScanner.Text()[5:]
	if err = procStat.Close(); err != nil {
		logger.ErrorLogger.Println(err)
		return nil, err
	}
	if err = procStatScanner.Err(); err != nil {
		logger.ErrorLogger.Println(err)
		return nil, err
	}
	cpuInfoSplit := strings.Fields(cpuInfo)
	idleTime, _ := strconv.ParseUint(cpuInfoSplit[3], 10, 64)
	totalTime := uint64(0)
	for _, s := range cpuInfoSplit {
		u, _ := strconv.ParseUint(s, 10, 64)
		totalTime += u
	}
	deltaIdleTime := idleTime - cpuUsage.PrevIdleTime
	deltaTotalTime := totalTime - cpuUsage.PrevTotalTime
	cpuUsage.CPUUsagePercentage = int64((1.0 - float64(deltaIdleTime)/float64(deltaTotalTime)) * 100.0)
	cpuUsage.PrevIdleTime = idleTime
	cpuUsage.PrevTotalTime = totalTime
	return cpuUsage, nil
}

func (m *CPUMemoryProfiler) monitorMemoryUsage() (*MemoryUsage, error) {
	var mStat runtime.MemStats
	runtime.ReadMemStats(&mStat)
	return &MemoryUsage{
		AllocatedHeap:  m.toMb(mStat.HeapAlloc),
		HeapInUse:      m.toMb(mStat.HeapInuse),
		TotalAllocated: m.toMb(mStat.TotalAlloc),
		SystemMemory:   m.toMb(mStat.Sys),
	}, nil
}

func (m *CPUMemoryProfiler) toMb(b uint64) int64 {
	return int64(b / 1048576)
}

func (m *CPUMemoryProfiler) SendCPUMemoryProfilingNode(stream protos.TransactionService_GetCPUMemoryProfilingResultServer) error {
	m.memoryRecordLock.Lock()
	defer m.memoryRecordLock.Unlock()
	var err error
	for _, profile := range m.CPUMemoryRecords {
		if err = stream.Send(profile); err != nil {
			break
		}
	}
	return err
}

func (m *CPUMemoryProfiler) SendCPUMemoryProfilingOrderer(stream protos.OrdererService_GetCPUMemoryProfilingResultServer) error {
	m.memoryRecordLock.Lock()
	defer m.memoryRecordLock.Unlock()
	var err error
	for _, profile := range m.CPUMemoryRecords {
		if err = stream.Send(profile); err != nil {
			break
		}
	}
	return err
}

func (m *CPUMemoryProfiler) SendCPUMemoryProfilingSequencer(stream protos.SequencerService_GetCPUMemoryProfilingResultServer) error {
	m.memoryRecordLock.Lock()
	defer m.memoryRecordLock.Unlock()
	var err error
	for _, profile := range m.CPUMemoryRecords {
		if err = stream.Send(profile); err != nil {
			break
		}
	}
	return err
}
