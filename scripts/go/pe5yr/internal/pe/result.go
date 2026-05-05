package pe

import "time"

// Result holds computed 5-year daily TTM P/E statistics for a single ticker.
type Result struct {
	Ticker          string
	StartDate       time.Time
	EndDate         time.Time
	MinPe           float64
	MinPeDate       time.Time
	MaxPe           float64
	MaxPeDate       time.Time
	Mean5yrPe       float64
	Ey5yrPe         float64
	Median5yrPe     float64
	ModePe          float64
	LatestPe        float64
	LastPrice       float64
	LastEps         float64
	Mean5yrEps      float64
	Shiller5yrPe    float64
	Profitable5yrPe float64
	Lossy5yrPe      float64
}
