package plot

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/BenjaminAlpert/go-stocks/internal/data"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func New(symbols []string, from time.Time, to time.Time, interval int) (io.WriterTo, error) {
	p := plot.New()

	prepPlot(p, interval)

	var lines []plotter.Line
	for _, symbol := range symbols {
		fmt.Printf("[INFO] Getting %s data from tiingo\n", symbol)
		entries, err := data.GetEnties(symbol, from, to, interval)
		if err != nil {
			return nil, err
		}
		fmt.Printf("[INFO] Done getting %s data from tiingo\n", symbol)

		line, err := makeLine(entries)
		if err != nil {
			return nil, err
		}
		lines = append(lines, *line)

	}

	err := doPlot(p, symbols, &lines)
	if err != nil {
		return nil, err
	}

	err = savePlot(p, "avg_rate.svg")
	if err != nil {
		return nil, err
	}

	writer, err := writePlot(p, "svg")
	if err != nil {
		return nil, err
	}

	return writer, nil
}

func prepPlot(p *plot.Plot, interval int) {
	p.Add(plotter.NewGrid())
	p.Title.Text = fmt.Sprintf("Normalized Rate of Change Over Time: (Prior 15 Day(s) Average Price) - (Prior %d Day(s) Average Price)) / (Prior %d Day(s) Average Price)", interval, interval)
	//p.Title.Text = fmt.Sprintf("Normalized Rate of Change Over Time: (Price - (Prior %d Day(s) Average Price)) / (Prior %d Day(s) Average Price)", interval, interval)
	p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02", Ticker: CustomTimeTicker{}}
	p.Y.Tick.Marker = CustomRateTicker{}
	p.X.Tick.Label.Rotation = math.Pi / 4
	p.X.Tick.Label.XAlign = draw.XRight
	p.X.Tick.Label.YAlign = draw.YCenter
}

func makeLine(entries []data.EntryType) (*plotter.Line, error) {
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

func writePlot(p *plot.Plot, outputformat string) (io.WriterTo, error) {
	writer, err := p.WriterTo(11*vg.Inch, 8*vg.Inch, outputformat)
	if err != nil {
		return nil, fmt.Errorf("unable to create io.WriterTo object for the %s plot, %s", outputformat, err)
	}
	return writer, nil
}
