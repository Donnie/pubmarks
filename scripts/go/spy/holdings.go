package main

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type holding struct {
	Name       string  `json:"name"`
	Ticker     string  `json:"ticker"`
	Weight     float64 `json:"weight"`
	SharesHeld float64 `json:"shares_held"`
}

var headerColumns = []string{"Name", "Ticker", "Weight", "Shares Held"}

func parseHoldings(xlsxPath string) ([]holding, error) {
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		return nil, fmt.Errorf("opening xlsx: %w", err)
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		return nil, fmt.Errorf("reading rows: %w", err)
	}

	headerRow, colIdx := findHeaderRow(rows, headerColumns)
	if headerRow < 0 {
		return nil, fmt.Errorf("header row not found (looking for %q)", headerColumns)
	}

	var holdings []holding
	for _, row := range rows[headerRow+1:] {
		if isEmptyRow(row) {
			break
		}
		weight, _ := strconv.ParseFloat(cell(row, colIdx[2]), 64)
		sharesHeld, _ := strconv.ParseFloat(cell(row, colIdx[3]), 64)
		holdings = append(holdings, holding{
			Name:       cell(row, colIdx[0]),
			Ticker:     cell(row, colIdx[1]),
			Weight:     weight,
			SharesHeld: sharesHeld,
		})
	}
	sort.Slice(holdings, func(i, j int) bool {
		return holdings[i].Weight > holdings[j].Weight
	})
	return holdings, nil
}

func findHeaderRow(rows [][]string, want []string) (int, []int) {
	for i, row := range rows {
		if idx := matchColumns(row, want); idx != nil {
			return i, idx
		}
	}
	return -1, nil
}

func matchColumns(row []string, want []string) []int {
	index := make(map[string]int, len(want))
	for i, w := range want {
		index[strings.ToLower(w)] = i
	}
	found := make([]int, len(want))
	for i := range found {
		found[i] = -1
	}
	for col, s := range row {
		key := strings.ToLower(strings.TrimSpace(s))
		if i, ok := index[key]; ok {
			found[i] = col
		}
	}
	if slices.Contains(found, -1) {
		return nil
	}
	return found
}

func isEmptyRow(row []string) bool {
	for _, s := range row {
		if strings.TrimSpace(s) != "" {
			return false
		}
	}
	return true
}

func cell(row []string, col int) string {
	if col < 0 || col >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[col])
}
