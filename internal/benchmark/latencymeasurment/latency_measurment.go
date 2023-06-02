package latencymeasurment

import "time"

type LatencyMeasurement func(startTime time.Time) func() (time.Time, time.Duration, time.Time)

type LatencyMeasurementInstance func() (time.Time, time.Duration, time.Time)

func Latency(startTime time.Time) func() (time.Time, time.Duration, time.Time) {
	return func() (time.Time, time.Duration, time.Time) {
		return startTime, time.Since(startTime), time.Now()
	}
}
