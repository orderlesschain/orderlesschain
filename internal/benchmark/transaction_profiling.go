package benchmark

type TransactionProfilingStat struct {
	transactionId          string
	sequenceDuration       int64
	orderDuration          int64
	endorseDuration        int64
	nodeCountEndorse       int
	endorseDurationAverage int64
	execDuration           int64
	nodeCountExec          int
	execDurationAverage    int64
	commitDuration         int64
	nodeCountCommit        int
	commitDurationAverage  int64
}

type TransactionProfilingStatSum struct {
	transactionProfiles map[string]*TransactionProfilingStat
}

func initTransactionProfilingStatSum() *TransactionProfilingStatSum {
	return &TransactionProfilingStatSum{transactionProfiles: map[string]*TransactionProfilingStat{}}
}
