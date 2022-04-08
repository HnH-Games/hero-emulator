package utils

import (
	rnd "crypto/rand"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

type Location struct {
	X float64
	Y float64
}

func (l *Location) String() string {
	return fmt.Sprintf("(%.1f,%.1f)", l.X, l.Y)
}

func CalculateDistance(loc1, loc2 *Location) float64 {
	return math.Sqrt(math.Abs(math.Pow(loc2.X-loc1.X, 2)) + math.Abs(math.Pow(loc2.Y-loc1.Y, 2)))
}

func RandInt(min, max int64) int64 {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			return
		}
	}()

	if min >= max {
		return min
	}

	d := rand.Int63n((max - min)) // [0, max-min)
	d += min                      // [min, max)
	return d
}

func RandFloat(min, max float64) float64 {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			return
		}
	}()

	f := rand.Float64() // [0.0f, 1.0f)
	f *= (max - min)    // (0.0f, max-min]
	f += min            // [min, max)
	return f
}

func RandFloats(min, max float64, count int) []float64 {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			return
		}
	}()

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	var arr []float64
	for ; count > 0; count-- {
		f := r.Float64() // [0.0f, 1.0f)
		f *= (max - min) // (0.0f, max-min]
		f += min         // [min, max)
		arr = append(arr, f)
	}

	return arr
}

func SigmaFunc(x float64) float64 {
	return 0.5*x*x + 10*x
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomInt(min, max int64) int64 {
	b := make([]byte, 8)
	_, err := rnd.Read(b)
	if err != nil {
		return 0
	}

	d := BytesToInt(b, false)
	if d < 0 {
		d *= -1
	}

	d %= max - min
	d += min

	return d
}

func generateRandomFloat(min, max float64) float64 {
	b := make([]byte, 8)
	_, err := rnd.Read(b)
	if err != nil {
		return 0
	}

	dRand := BytesToInt(b, false)
	if dRand < 0 {
		dRand *= -1
	}

	dMin := BytesToInt(FloatToBytes(min, 8, false), false)
	dMax := BytesToInt(FloatToBytes(max, 8, false), false)
	dRand %= (dMax - dMin)
	dRand += dMin

	return BytesToFloat(IntToBytes(uint64(dRand), 8, false), false)
}

func PvPFunc(val int) int {
	v := math.Sqrt(float64(val))
	pow := 4 * math.E / 5
	v = math.Pow(v, pow)
	return int(v)
}

func ParseFloat(d string) float64 {
	f, err := strconv.ParseFloat(d, 64)
	if err != nil {
		return 0
	}

	return f
}
