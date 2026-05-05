package pubmarks

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"donnie.in/sniper360/apps/pubmarks/internal/cdn"
	"donnie.in/sniper360/apps/pubmarks/internal/pe"
)

func init() {
	// Register the HTTP function for the Functions Framework (2nd gen / Cloud Run).
	functions.HTTP("AvgPe", AvgPe)
}

// AvgPe is an HTTP Cloud Function entrypoint for GCP.
// It expects a query parameter "ticker" and returns JSON matching the previous Gin handler.
func AvgPe(w http.ResponseWriter, r *http.Request) {
	ticker := r.URL.Query().Get("ticker")
	if ticker == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "missing required query param: ticker (example: /avgpe?ticker=MSFT)",
		})
		return
	}

	res, err := pe.FiveYearAveragePe(ticker, time.Now())
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, cdn.ErrNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, ToPayload(res))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Payload is the successful JSON body for AvgPe (and local CLI output).
type Payload struct {
	Ticker          string  `json:"ticker"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	MinPe           float64 `json:"p_e_min"`
	MinPeDate       string  `json:"p_e_min_date"`
	MaxPe           float64 `json:"p_e_max"`
	MaxPeDate       string  `json:"p_e_max_date"`
	Mean5yrPe       float64 `json:"p_e_mean_5yr"`
	Median5yrPe     float64 `json:"p_e_median_5yr"`
	Mode5yrPe       float64 `json:"p_e_mode_5yr"`
	Avg5yrPe        float64 `json:"p_e_avg_5yr"`
	Ey5yrPe         float64 `json:"p_e_earningsyield_5yr"`
	LatestPe        float64 `json:"p_e_last"`
	Shiller5yrPe    float64 `json:"p_e_shiller_5yr"`
	Profitable5yrPe float64 `json:"p_e_profitable_5yr"`
	Lossy5yrPe      float64 `json:"p_e_lossy_5yr"`
	LastPrice       float64 `json:"price_last"`
	LastEps         float64 `json:"eps_last"`
}

// ToPayload converts pe.Result into the JSON output contract.
// Dates are ISO-8601 strings; floats are rounded to match precision expectations.
func ToPayload(r pe.Result) Payload {
	return Payload{
		Ticker:          r.Ticker,
		StartDate:       r.StartDate.Format("2006-01-02"),
		EndDate:         r.EndDate.Format("2006-01-02"),
		MinPe:           round4(r.MinPe),
		MinPeDate:       r.MinPeDate.Format("2006-01-02"),
		MaxPe:           round4(r.MaxPe),
		MaxPeDate:       r.MaxPeDate.Format("2006-01-02"),
		Mean5yrPe:       round4(r.Mean5yrPe),
		Median5yrPe:     round4(r.Median5yrPe),
		Mode5yrPe:       round4(math.Round(r.ModePe)),
		Avg5yrPe:        round4(r.Mean5yrPe), // alias of p_e_mean_5yr
		Ey5yrPe:         round4(r.Ey5yrPe),
		LatestPe:        round4(r.LatestPe),
		Shiller5yrPe:    round4(r.Shiller5yrPe),
		Profitable5yrPe: round4(r.Profitable5yrPe),
		Lossy5yrPe:      round4(r.Lossy5yrPe),
		LastPrice:       round2(r.LastPrice),
		LastEps:         round4(r.LastEps),
	}
}

func round4(v float64) float64 { return math.Round(v*10000) / 10000 }
func round2(v float64) float64 { return math.Round(v*100) / 100 }
