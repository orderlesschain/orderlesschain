package crdtmanagerv2

import (
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"hash/adler32"
	"strconv"
	"strings"
)

type HappenedBeforeType int

const (
	NoHappenedBefore          HappenedBeforeType = -1
	FirstHappenedBeforeSecond HappenedBeforeType = 1
	SecondHappenedBeforeFirst HappenedBeforeType = 0
)

func NewClockWithCounter(seed string, counter int64) *protos.CRDTClock {
	return &protos.CRDTClock{
		Seed:  GetClockSeeId(seed),
		Count: counter,
	}
}

func GetClockSeeId(id string) int64 {
	return int64(adler32.Checksum([]byte(id)))
}

func StringToClock(s string) (*protos.CRDTClock, error) {
	c := &protos.CRDTClock{}
	str := strings.Split(s, ".")
	count, err := strconv.Atoi(str[0])
	if err != nil {
		return c, err
	}
	seed, err := strconv.Atoi(str[1])
	if err != nil {
		return c, err
	}

	c.Count = int64(count)
	c.Seed = int64(seed)
	return c, nil
}

func ClockToString(c *protos.CRDTClock) string {
	cnt := strconv.FormatInt(c.Count, int(protos.TimeBase_BASE))
	sd := strconv.FormatInt(c.Seed, int(protos.TimeBase_BASE))
	return cnt + "." + sd
}

func ClockHappenedBefore(first, second *protos.CRDTClock) HappenedBeforeType {
	if first.Seed != second.Seed {
		// No happened-before
		return NoHappenedBefore
	}
	if first.Count < second.Count {
		// First happened-before second
		return FirstHappenedBeforeSecond
	}
	if first.Count > second.Count {
		// Second happened-before first
		return SecondHappenedBeforeFirst
	}
	// No happened-before
	return NoHappenedBefore
}
