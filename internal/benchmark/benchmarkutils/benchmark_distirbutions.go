package benchmarkutils

import (
	wr "github.com/mroth/weightedrand"
	"math/rand"
	"time"
)

type BenchmarkUtils struct {
	defaultKeyIndex               int
	numberOfKeys                  int
	numberOfKeysMean              float64
	numberOfKeysStandardDeviation float64
	counter                       int
	baseIndex                     int

	zipf   *rand.Zipf
	random *rand.Rand

	read0write100 *wr.Chooser
	read10write90 *wr.Chooser
	read20write80 *wr.Chooser
	read30write70 *wr.Chooser
	read40write60 *wr.Chooser
	read50write50 *wr.Chooser
	read60write40 *wr.Chooser
	read70write30 *wr.Chooser
	read80write20 *wr.Chooser
	read90write10 *wr.Chooser
	read100write0 *wr.Chooser

	conflict0uniform100 *wr.Chooser
	conflict10uniform90 *wr.Chooser
	conflict20uniform80 *wr.Chooser
	conflict30uniform70 *wr.Chooser
	conflict40uniform60 *wr.Chooser
	conflict50uniform50 *wr.Chooser
	conflict60uniform40 *wr.Chooser
	conflict70uniform30 *wr.Chooser
	conflict80uniform20 *wr.Chooser
	conflict90uniform10 *wr.Chooser
	conflict100uniform0 *wr.Chooser
}

func NewBenchmarkDist(defaultKeyIndex, currentClientId, totalTransactionsCount int, numberOfKeys int64) *BenchmarkUtils {
	tempUtils := &BenchmarkUtils{
		defaultKeyIndex:               defaultKeyIndex,
		baseIndex:                     currentClientId * totalTransactionsCount,
		numberOfKeys:                  int(numberOfKeys),
		numberOfKeysMean:              float64(numberOfKeys / 2),
		numberOfKeysStandardDeviation: float64(numberOfKeys) * .15,

		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		zipf: rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())),
			1.5, float64((numberOfKeys)/10),
			uint64(numberOfKeys)),
	}
	tempUtils.read0write100, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 0},
		wr.Choice{Item: false, Weight: 100},
	)
	tempUtils.read10write90, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 10},
		wr.Choice{Item: false, Weight: 90},
	)
	tempUtils.read20write80, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 20},
		wr.Choice{Item: false, Weight: 80},
	)
	tempUtils.read30write70, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 30},
		wr.Choice{Item: false, Weight: 70},
	)
	tempUtils.read40write60, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 40},
		wr.Choice{Item: false, Weight: 60},
	)
	tempUtils.read50write50, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 50},
		wr.Choice{Item: false, Weight: 50},
	)
	tempUtils.read60write40, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 60},
		wr.Choice{Item: false, Weight: 40},
	)
	tempUtils.read70write30, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 70},
		wr.Choice{Item: false, Weight: 30},
	)
	tempUtils.read80write20, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 80},
		wr.Choice{Item: false, Weight: 20},
	)
	tempUtils.read90write10, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 90},
		wr.Choice{Item: false, Weight: 10},
	)
	tempUtils.read100write0, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 100},
		wr.Choice{Item: false, Weight: 0},
	)

	tempUtils.conflict0uniform100, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 0},
		wr.Choice{Item: false, Weight: 100},
	)
	tempUtils.conflict10uniform90, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 10},
		wr.Choice{Item: false, Weight: 90},
	)
	tempUtils.conflict20uniform80, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 20},
		wr.Choice{Item: false, Weight: 80},
	)
	tempUtils.conflict30uniform70, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 30},
		wr.Choice{Item: false, Weight: 70},
	)
	tempUtils.conflict40uniform60, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 40},
		wr.Choice{Item: false, Weight: 60},
	)
	tempUtils.conflict50uniform50, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 50},
		wr.Choice{Item: false, Weight: 50},
	)
	tempUtils.conflict60uniform40, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 60},
		wr.Choice{Item: false, Weight: 40},
	)
	tempUtils.conflict70uniform30, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 70},
		wr.Choice{Item: false, Weight: 30},
	)
	tempUtils.conflict80uniform20, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 80},
		wr.Choice{Item: false, Weight: 20},
	)
	tempUtils.conflict90uniform10, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 90},
		wr.Choice{Item: false, Weight: 10},
	)
	tempUtils.conflict100uniform0, _ = wr.NewChooser(
		wr.Choice{Item: true, Weight: 100},
		wr.Choice{Item: false, Weight: 0},
	)

	return tempUtils
}

func (b *BenchmarkUtils) GetZipDistributionKeyIndex() int {
	return int(b.zipf.Uint64())
}

func (b *BenchmarkUtils) GetZipDistributionKeyIndexWithDifferentNumberOfKeys(numberOfKeys int) int {
	zipf := rand.NewZipf(rand.New(rand.NewSource(time.Now().UnixNano())),
		1.5, float64((numberOfKeys)/10),
		uint64(numberOfKeys))
	return int(zipf.Uint64())
}

func (b *BenchmarkUtils) GetUniformDistributionKeyIndex() int {
	return b.random.Intn(b.numberOfKeys)
}

func (b *BenchmarkUtils) GetUniformDistributionKeyIndexWithDifferentNumberOfKeys(numberOfKeys int) int {
	return b.random.Intn(numberOfKeys)
}

func (b *BenchmarkUtils) GetRandomFloatWithMinAndMax(min, max float64) float64 {
	return min + b.random.Float64()*(max-min)
}

func (b *BenchmarkUtils) GetGaussianDistributionKeyIndex() int {
	randomValue := int(b.random.NormFloat64()*b.numberOfKeysStandardDeviation + b.numberOfKeysMean)
	if randomValue < 0 {
		return 0
	}
	if randomValue >= b.numberOfKeys {
		return b.numberOfKeys - 1
	}
	return randomValue
}

func (b *BenchmarkUtils) GetGaussianDistributionKeyIndexWithDifferentNumberOfKey(numberOfKeys int) int {
	numberOfKeysMean := float64(numberOfKeys / 2)
	numberOfKeysStandardDeviation := float64(numberOfKeys) * .15
	randomValue := int(b.random.NormFloat64()*numberOfKeysStandardDeviation + numberOfKeysMean)
	if randomValue < 0 {
		return 0
	}
	if randomValue >= numberOfKeys {
		return numberOfKeys - 1
	}
	return randomValue
}

func (b *BenchmarkUtils) Read0Write100() (read bool) {
	return b.read0write100.Pick().(bool)
}

func (b *BenchmarkUtils) Read10Write90() (read bool) {
	return b.read10write90.Pick().(bool)
}

func (b *BenchmarkUtils) Read20Write80() (read bool) {
	return b.read20write80.Pick().(bool)
}

func (b *BenchmarkUtils) Read30Write70() (read bool) {
	return b.read30write70.Pick().(bool)
}

func (b *BenchmarkUtils) Read40Write60() (read bool) {
	return b.read40write60.Pick().(bool)
}

func (b *BenchmarkUtils) Read50Write50() (read bool) {
	return b.read50write50.Pick().(bool)
}

func (b *BenchmarkUtils) Read60Write40() (read bool) {
	return b.read60write40.Pick().(bool)
}

func (b *BenchmarkUtils) Read70Write30() (read bool) {
	return b.read70write30.Pick().(bool)
}

func (b *BenchmarkUtils) Read80Write20() (read bool) {
	return b.read80write20.Pick().(bool)
}

func (b *BenchmarkUtils) Read90Write10() (read bool) {
	return b.read90write10.Pick().(bool)
}

func (b *BenchmarkUtils) Read100Write0() (read bool) {
	return b.read100write0.Pick().(bool)
}

func (b *BenchmarkUtils) Conflict0Uniform100() int {
	if b.conflict0uniform100.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict10Uniform90() int {
	if b.conflict10uniform90.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict20Uniform80() int {
	if b.conflict20uniform80.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict30Uniform70() int {
	if b.conflict30uniform70.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict40Uniform60() int {
	if b.conflict40uniform60.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict50Uniform50() int {
	if b.conflict50uniform50.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict60Uniform40() int {
	if b.conflict60uniform40.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict70Uniform30() int {
	if b.conflict70uniform30.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict80Uniform20() int {
	if b.conflict80uniform20.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict90Uniform10() int {
	if b.conflict90uniform10.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}

func (b *BenchmarkUtils) Conflict100Uniform0() int {
	if b.conflict100uniform0.Pick().(bool) {
		return b.defaultKeyIndex
	} else {
		b.counter++
		return b.baseIndex + b.counter
	}
}
