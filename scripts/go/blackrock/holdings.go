package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

type holding struct {
	Name       string  `json:"name"`
	Ticker     string  `json:"ticker"`
	Weight     float64 `json:"weight"`
	SharesHeld float64 `json:"shares_held"`
}

type holdingsPayload struct {
	AaData [][]json.RawMessage `json:"aaData"`
}

func parseHoldingsJSON(raw []byte) ([]holding, error) {
	var p holdingsPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("unmarshal holdings json: %w", err)
	}

	var out []holding
	for _, row := range p.AaData {
		h, err := rowToHolding(row)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func rowToHolding(row []json.RawMessage) (holding, error) {
	if len(row) < 8 {
		return holding{}, fmt.Errorf("holdings row too short: %d cols", len(row))
	}

	ticker := jsonString(row[0])
	name := jsonString(row[1])
	weight, err := rawNumber(row[5])
	if err != nil {
		return holding{}, fmt.Errorf("weight: %w", err)
	}
	shares, err := rawNumber(row[7])
	if err != nil {
		return holding{}, fmt.Errorf("shares: %w", err)
	}

	return holding{
		Name:       name,
		Ticker:     ticker,
		Weight:     weight,
		SharesHeld: shares,
	}, nil
}

func jsonString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return ""
}

type displayRaw struct {
	Raw float64 `json:"raw"`
}

func rawNumber(raw json.RawMessage) (float64, error) {
	var n float64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	var dr displayRaw
	if err := json.Unmarshal(raw, &dr); err == nil {
		return dr.Raw, nil
	}
	return 0, fmt.Errorf("not a number or display/raw object: %s", string(raw))
}
