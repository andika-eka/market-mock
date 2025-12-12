package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"
)

type CandleSeries struct {
	Symbol  string
	Candles []Candle
}

func NewCandleSeries(symbol string, size int) *CandleSeries {
	return &CandleSeries{
		Symbol:  symbol,
		Candles: make([]Candle, 0, size)}
}

func (c *CandleSeries) AddCandle(candle Candle) {
	c.Candles = append(c.Candles, candle)
}

func (c Candle) ToCSV() []string {
	return []string{
		c.Time.Format("2006-01-02 15:04:05"),
		fmt.Sprintf("%.2f", c.Open),
		fmt.Sprintf("%.2f", c.High),
		fmt.Sprintf("%.2f", c.Low),
		fmt.Sprintf("%.2f", c.Close),
	}
}

func (c CandleSeries) ToCSV(header bool) [][]string {
	rows := make([][]string, 0, len(c.Candles))
	if header {
		rows = append(rows, []string{"Symbol", "Time", "Open", "High", "Low", "Close"})
	}

	for _, candle := range c.Candles {
		candleCSV := candle.ToCSV()
		rows = append(rows, append([]string{c.Symbol}, candleCSV...))
	}
	return rows
}

func getDataSize(start time.Time, end time.Time, interval time.Duration) int {
	duration := end.Sub(start)
	return int(duration / interval)
}

func GetDataSeries(symbol string, start time.Time, end time.Time, interval time.Duration) *CandleSeries {
	size := getDataSize(start, end, interval)
	series := NewCandleSeries(symbol, size)
	for t := start; t.Before(end); t = t.Add(interval) {
		candle := GetOHLC(symbol, t, interval)
		series.AddCandle(candle)
	}
	return series
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

	//get last 10 years of data
	endDate := time.Now()
	startDate := endDate.AddDate(-10, 0, 0)
	interval := 300 * time.Hour
	header := true
	for _, s := range symbols {
		series := GetDataSeries(s, startDate, endDate, interval)
		if header {
			rows := series.ToCSV(true)
			header = false
			for _, row := range rows {
				writer.Write(row)
			}
		} else {
			rows := series.ToCSV(false)
			for _, row := range rows {
				writer.Write(row)
			}
		}
	}
}
