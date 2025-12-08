package main

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"time"
)

const (
	MinPrice               = 1.0
	MaxPrice               = 200000.0
	MinVol                 = 0.01
	MaxVol                 = 0.40
	MultiplierInterval     = 86400 * 7
	MinBasePriceMultplier  = 0.2
	MaxBasePriceMultplier  = 2
	MinVolatilityMultplier = 0.2
	MaxVolatilityMultplier = 2
)

// https://en.wikipedia.org/wiki/Smoothstep
func smoothStep(t float64) float64 {
	return t * t * (3 - 2*t)
}

func getHashFloat(seed uint64) float64 {
	r := rand.New(rand.NewPCG(seed, seed^0x5CA1AB1E))
	return r.Float64()
}

func GetMultiplier(symbol string, t time.Time) (price float64, volatility float64) {

	unix := t.Unix()
	// get week number x since unix 0
	currentStep := unix / MultiplierInterval
	nextStep := currentStep + 1 //next week

	//progress within week - how long since week start
	progress := float64(unix%MultiplierInterval) / float64(MultiplierInterval)

	progress = smoothStep(progress) // TODO check when turned off

	hashSymbol := fnv.New64a()
	hashSymbol.Write([]byte(symbol))

	binary.Write(hashSymbol, binary.BigEndian, currentStep)
	seedStart := hashSymbol.Sum64()

	//multiplier at start
	startPriceMult := MinBasePriceMultplier + getHashFloat(seedStart)*(MaxBasePriceMultplier-MinBasePriceMultplier)
	startVolMult := MinBasePriceMultplier + getHashFloat(seedStart+1)*(MaxVolatilityMultplier-MinVolatilityMultplier)

	hashSymbol.Reset()
	hashSymbol.Write([]byte(symbol))
	binary.Write(hashSymbol, binary.BigEndian, nextStep)
	seedEnd := hashSymbol.Sum64()

	//multiplier at start
	endPriceMult := MinBasePriceMultplier + getHashFloat(seedEnd)*(MaxBasePriceMultplier-MinBasePriceMultplier)
	endVolMult := MinBasePriceMultplier + getHashFloat(seedEnd+1)*(MaxVolatilityMultplier-MinVolatilityMultplier)

	//multiplier current - interpolation
	price = startPriceMult + (endPriceMult-startPriceMult)*progress
	volatility = startVolMult + (endVolMult-startVolMult)*progress

	return
}

func GetPrice(symbol string, t time.Time) float64 {

	hashSymbol := fnv.New64a()
	hashSymbol.Write([]byte(symbol))
	seedSymbol := hashSymbol.Sum64()
	valueSymbol := rand.New(rand.NewPCG(seedSymbol, seedSymbol^0x5CA1AB1E))
	symbolBasePrice := MinPrice + valueSymbol.Float64()*(MaxPrice-MinPrice)
	symbolVolatility := MinVol + valueSymbol.Float64()*(MaxVol-MinVol)

	priceMulti, volMulti := GetMultiplier(symbol, t)
	symbolBasePrice = symbolBasePrice * priceMulti
	symbolVolatility = symbolVolatility * volMulti

	timeSeed := uint64(t.Unix() / 10)
	hashTime := fnv.New64a()
	binary.Write(hashTime, binary.BigEndian, seedSymbol)
	binary.Write(hashTime, binary.BigEndian, timeSeed)

	seedFinal := hashTime.Sum64()
	valueTime := rand.New(rand.NewPCG(seedFinal, seedFinal^0x5CA1AB1E))

	variation := (valueTime.Float64() * 2 * symbolVolatility) - symbolVolatility
	price := symbolBasePrice * (1 + variation)
	return price

}

func main() {
	symbol := "BTC-USD"
	targetTime := time.Date(2025, 12, 8, 10, 0, 0, 0, time.UTC)

	fmt.Printf(" %s - %s\n", symbol, targetTime)

	price1 := GetPrice(symbol, targetTime)

	price2 := GetPrice(symbol, targetTime)
	fmt.Printf("result: %.4f\n - %.4f\n", price1, price2)

	for i := 0; i < 1000; i += 10 {
		fmt.Printf("%.4f\n", GetPrice(symbol, targetTime.Add(time.Duration(i)*time.Second)))
	}
}
