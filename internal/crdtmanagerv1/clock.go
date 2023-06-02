package crdtmanagerv1

import (
	"gitlab.lrz.de/orderless/orderlesschain/protos/goprotos"
	"hash/adler32"
	"strconv"
	"strings"
)

// NewClock Initializes a clock. The `seed` is a string which uniquely identifies the
// clock in the network
func NewClock(seed string) *protos.Clock {
	return &protos.Clock{
		Seed:  GetClockSeeId(seed),
		Count: 1,
	}
}

func GetClockSeeId(id string) int64 {
	return int64(adler32.Checksum([]byte(id)))
}

// Tick Increments Clock count
func Tick(c *protos.Clock) {
	c.Count++
}

// Timestamp Returns Timestamp that uniquely identifies the state (clock and count) in the
// network
func Timestamp(c *protos.Clock) string {
	return c.String()
}

// UpdateClock Updates a Clock based on another clock or string representation. If the
// current Clock count.seed value is higher, no changes are done. Otherwise, the
// clock updates to the upper count
func UpdateClock(c *protos.Clock, rcv interface{}) error {
	var err error
	rcvC := &protos.Clock{}
	switch t := rcv.(type) {
	case protos.Clock:
		rcvC = &t

	case string:
		rcvC, err = strToClock(t)
		if err != nil {
			return err
		}
	}

	rcvCan, err := canonical(ConvertClockToString(rcvC))
	if err != nil {
		return err
	}

	currCan, err := canonical(ConvertClockToString(c))
	if err != nil {
		return err
	}

	if rcvCan > currCan {
		c.Count = rcvC.Count
	}
	return nil
}

// Returns the canonical value of clock. The canonical value of the logical
// clock is a float64 type in the form of <Clock.count>.<Clock.seed>. The
// Clock.seed value must be unique per Clock in the network.
func canonical(c string) (float64, error) {
	fc, err := strconv.ParseFloat(c, 10)
	return fc, err
}

func strToClock(s string) (*protos.Clock, error) {
	c := &protos.Clock{}
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

func ConvertClockToString(c *protos.Clock) string {
	cnt := strconv.FormatInt(c.Count, int(protos.TimeBase_BASE))
	sd := strconv.FormatInt(c.Seed, int(protos.TimeBase_BASE))
	return cnt + "." + sd
}
