package plot

import (
	"fmt"
	"time"

	"gonum.org/v1/plot"
)

type CustomTimeTicker struct{}
type CustomRateTicker struct{}

func (t CustomTimeTicker) Ticks(min, max float64) []plot.Tick {
	var ticks []plot.Tick
	start := time.Unix(int64(min), 0)
	end := time.Unix(int64(max), 0)

	for start.Before(end) {
		//numYears := end.Sub(start).Hours() / 24 / 365
		value := float64(start.Unix())
		ticks = append(ticks, plot.Tick{Value: value, Label: fmt.Sprintf("%f", value)})
		// if numYears <= 1 {
		// 	start = start.AddDate(0, 2, 0)
		// } else {
		start = start.AddDate(1, 0, 0)
		//}
	}
	ticks = append(ticks, plot.Tick{Value: max, Label: fmt.Sprintf("%f", max)})
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
