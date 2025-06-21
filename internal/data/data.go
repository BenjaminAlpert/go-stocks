package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	tokenEnvName = "TIINGO_TOKEN"
)

type EntryType struct {
	Close   float64   `json:"close"`
	Date    time.Time `json:"date"`
	AvgRate float64   `json:"rate"`
}

func GetEnties(symbol string, from, to time.Time, lookBackInterval int) ([]EntryType, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	type rawEntriesType struct {
		Close float64 `json:"close"`
		Date  string  `json:"date"`
	}

	type rawDetailsResponseType struct {
		Detail string `json:"detail"`
	}

	var rawEntries []rawEntriesType

	url := fmt.Sprintf(
		"https://api.tiingo.com/tiingo/daily/%s/prices?startDate=%s&endDate=%s",
		strings.TrimSpace(strings.Replace(symbol, "/", "-", -1)),
		url.QueryEscape(from.Format("2006-1-2")),
		url.QueryEscape(to.Format("2006-1-2")),
	)

	const clientTimeout = 10 * time.Second
	client := &http.Client{Timeout: clientTimeout}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("tiingo error: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		contents, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(contents, &rawEntries)
		if err != nil {
			return nil, fmt.Errorf("tiingo error: %s", err)
		}
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tiingo error: %s", err)
	} else {
		contents, _ := io.ReadAll(resp.Body)
		var rawDetail rawDetailsResponseType
		err = json.Unmarshal(contents, &rawDetail)
		if err != nil {
			return nil, fmt.Errorf("unable to parse tiingo error message (%s): %s", contents, err)
		}
		return nil, fmt.Errorf("tiingo error: %s", rawDetail.Detail)
	}

	var entries []EntryType

	for _, rawEntry := range rawEntries {
		date, err := time.Parse("2006-01-02", rawEntry.Date[0:10])
		if err != nil {
			return nil, fmt.Errorf("unable to parse date from tiingo response: %s", err)
		}
		entries = append(entries,
			EntryType{
				Close: rawEntry.Close,
				Date:  date,
			},
		)
	}

	entries = addAvgRateToEntries(entries, lookBackInterval)

	return entries, nil
}

func getToken() (string, error) {
	token := os.Getenv(tokenEnvName)
	if token == "" {
		return "", fmt.Errorf("tiingo token not found in env var %s", tokenEnvName)
	}
	return token, nil
}

func addAvgRateToEntries(inEntries []EntryType, lookBackInterval int) []EntryType {
	var outEntries []EntryType

	for index := range inEntries {
		if index >= lookBackInterval {
			avg := average(inEntries[index-lookBackInterval : index])
			avgRate := (average(inEntries[index-15:index]) - avg) / avg
			// avgRate := (inEntries[index].Close - avg) / avg

			outEntries = append(outEntries, EntryType{
				Close:   inEntries[index].Close,
				AvgRate: avgRate,
				Date:    inEntries[index].Date,
			})
		}
	}
	return outEntries
}

func average(entries []EntryType) float64 {
	sum := float64(0)
	for _, value := range entries {
		sum += value.Close
	}
	return sum / float64(len(entries))
}
