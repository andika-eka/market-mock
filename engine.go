package main

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"hash/fnv"
	"log"
	"math/rand/v2"
	"os"
	"time"
)

const (
	MinPrice               = 1.0
	MaxPrice               = 200000.0
	MinVol                 = 0.05
	MaxVol                 = 0.30
	MultiplierInterval     = 86400 * 600
	MinBasePriceMultplier  = 0.2
	MaxBasePriceMultplier  = 1.5
	MinVolatilityMultplier = 0.2
	MaxVolatilityMultplier = 0.9
	dataPoints             = 120
)

type Candle struct {
	Symbol string    `json:"symbol"`
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
}

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

func GetOHLC(symbol string, start time.Time, duration time.Duration) Candle {

	stepSize := duration / time.Duration(dataPoints-1)

	var prices [dataPoints]float64

	for i := 0; i < dataPoints; i++ {
		sampleTime := start.Add(time.Duration(i) * stepSize)
		prices[i] = GetPrice(symbol, sampleTime)
	}

	high := prices[0]
	low := prices[0]

	for _, p := range prices {
		if p > high {
			high = p
		}
		if p < low {
			low = p
		}
	}

	// Return the clean struct
	return Candle{
		Symbol: symbol,
		Time:   start,
		Open:   prices[0],
		High:   high,
		Low:    low,
		Close:  prices[dataPoints-1],
	}
}

func main() {
	symbols := []string{"BTC-USD", "ETH-USD", "STABLE-COIN", "GOLD", "PLATINUM", "SILVER"}
	filename := "market_data.csv"

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"symbol", "date", "open", "high", "low", "close"})

	//get last 10 years of data
	endDate := time.Now()
	startDate := endDate.AddDate(-10, 0, 0)
	interval := 300 * time.Hour

	for t := startDate; t.Before(endDate); t = t.Add(interval) {
		for _, s := range symbols {
			c := GetOHLC(s, t, interval)
			row := []string{
				c.Symbol,
				c.Time.Format("2006-01-02 15:04:05"),
				fmt.Sprintf("%.2f", c.Open),
				fmt.Sprintf("%.2f", c.High),
				fmt.Sprintf("%.2f", c.Low),
				fmt.Sprintf("%.2f", c.Close),
			}
			writer.Write(row)
		}
	}
}
