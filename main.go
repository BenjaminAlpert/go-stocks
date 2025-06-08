package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dvcrn/go-1password-cli/op"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type entryType struct {
	Close   float64   `json:"close"`
	Date    time.Time `json:"date"`
	AvgRate float64   `json:"rate"`
	Symbol  string    `json:"symbol"`
}

const (
	onePasswordItemId = "z35jsmkboau56wxn3ftgqpi66y"
	period            = 365 * 20 // number of days to show
	lookBackInterval  = 365 * 1  // number of days before date index to average over
)

func main() {

	token, err := getTokenFromOnePassword()
	if err != nil {
		panic(err)
	}

	now := time.Now()
	to := now
	from := now.Add(-time.Duration((period + lookBackInterval) * time.Hour * 24))

	symbols := []string{"dia", "spy", "vt"}

	p := plot.New()
	prepPlot(p)

	var lines []plotter.Line

	for _, symbol := range symbols {
		entries, err := newQuote(symbol, token, from, to)
		if err != nil {
			panic(err)
		}
		entries = processEntries(entries)

		line, err := makeLine(entries)
		if err != nil {
			panic(err)
		}
		lines = append(lines, *line)

	}

	err = doPlot(p, symbols, &lines)
	if err != nil {
		panic(err)
	}

	err = savePlot(p, "avg_rate.svg")
	if err != nil {
		panic(err)
	}

}

func processEntries(inEntries []entryType) []entryType {

	var outEntries []entryType

	for index := range inEntries {
		if index >= lookBackInterval {
			avgRate := (inEntries[index].Close - average(inEntries[index-lookBackInterval:index])) / inEntries[index].Close

			outEntries = append(outEntries, entryType{
				AvgRate: avgRate,
				Date:    inEntries[index].Date,
			})
		}
	}
	return outEntries
}

func average(entries []entryType) float64 {
	sum := float64(0)
	for _, value := range entries {
		sum += value.Close
	}
	return sum / float64(len(entries))
}

type CustomTimeTicker struct{}
type CustomRateTicker struct {
}

func (t CustomTimeTicker) Ticks(min, max float64) []plot.Tick {
	var ticks []plot.Tick
	start := time.Unix(int64(min), 0)
	end := time.Unix(int64(max), 0)

	for start.Before(end) {
		numYears := math.Ceil(end.Sub(start).Hours() / 24 / 365)
		value := float64(start.Unix())
		ticks = append(ticks, plot.Tick{Value: value, Label: fmt.Sprintf("%f", value)})
		if numYears == 1 {
			start = start.AddDate(0, 1, 0)
		} else {
			start = start.AddDate(1, 0, 0)
		}
	}
	return ticks
}

func (t CustomRateTicker) Ticks(min, max float64) []plot.Tick {

	ticks := plot.DefaultTicks.Ticks(plot.DefaultTicks{}, min, max)
	for index := range ticks {
		ticks[index].Label = fmt.Sprintf("%02d%%", int(ticks[index].Value*100))
	}
	ticks = append(ticks, plot.Tick{Value: float64(0), Label: "00%"})

	return ticks
}

func prepPlot(p *plot.Plot) {
	p.Add(plotter.NewGrid())
	p.Title.Text = fmt.Sprintf("Normalized Rate of Change Over Time: (Price - Prior %d Day(s) Average Price) / Price", lookBackInterval)

	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02", Ticker: CustomTimeTicker{}}
	p.Y.Tick.Marker = CustomRateTicker{}
}

func makeLine(entries []entryType) (*plotter.Line, error) {
	pts := make(plotter.XYs, len(entries))
	for index, entry := range entries {
		pts[index].X = float64(entry.Date.Unix())
		pts[index].Y = entry.AvgRate
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		return nil, fmt.Errorf("unable to make line plot, %s", err)
	}
	return line, nil
}

func doPlot(p *plot.Plot, symbols []string, lines *[]plotter.Line) error {
	var vs []any
	for index, symbol := range symbols {
		vs = append(vs, symbol, (*lines)[index])
	}
	err := plotutil.AddLines(p, vs...)
	if err != nil {
		return fmt.Errorf("unable to add line plot to plot, %s", err)
	}
	return nil
}

func savePlot(p *plot.Plot, outputPath string) error {
	err := p.Save(11*vg.Inch, 8*vg.Inch, outputPath)
	if err != nil {
		return fmt.Errorf("unable to save plot to image file (%s), %s", outputPath, err)
	}
	return nil
}

func newQuote(symbol string, token string, from time.Time, to time.Time) ([]entryType, error) {
	type tQuoteType struct {
		Open  float64 `json:"open"`
		Close float64 `json:"close"`
		Date  string  `json:"date"`
	}

	var tquotes []tQuoteType

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
		err = json.Unmarshal(contents, &tquotes)
		if err != nil {
			return nil, fmt.Errorf("tiingo error: %s", err)
		}
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tiingo error: %s", err)
	} else {
		contents, _ := io.ReadAll(resp.Body)
		type detail struct {
			Detail string `json:"detail"`
		}
		var d detail
		err = json.Unmarshal(contents, &d)
		if err != nil {
			return nil, fmt.Errorf("unable to parse tiingo error message (%s): %s", contents, err)
		}
		return nil, fmt.Errorf("tiingo error: %s", d.Detail)
	}

	var entries []entryType

	for _, tquote := range tquotes {
		date, err := time.Parse("2006-01-02", tquote.Date[0:10])
		if err != nil {
			return nil, fmt.Errorf("unable to parse date from tiingo response: %s", err)
		}
		entries = append(entries,
			entryType{
				Close: tquote.Close,
				Date:  date,
			},
		)
	}

	return entries, nil

}

func getTokenFromOnePassword() (string, error) {
	client := op.NewOpClient()
	item, err := client.Item(onePasswordItemId)
	if err != nil {
		return "", fmt.Errorf("unable to get tiingo token from 1password")
	}
	for _, field := range item.Fields {
		if field.Label == "token" {
			return field.Value, nil
		}
	}
	return "", errors.New("unable to find token field from 1password item")
}
