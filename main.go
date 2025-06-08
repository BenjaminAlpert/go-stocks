package main

import (
	"fmt"
	"io"
	"time"

	"github.com/BenjaminAlpert/go-stocks/internal/plot"
	"github.com/BenjaminAlpert/go-stocks/internal/server"
)

const (
	period           = 365 * 20 // number of days to show
	lookBackInterval = 365 * 1  // number of days before date index to average over
	cacheTimeout     = time.Minute
)

var (
	Writer io.WriterTo
)

func main() {

	symbols := []string{"dia", "spy", "vt"}

	ticker := time.NewTicker(cacheTimeout)
	quit := make(chan struct{})

	generateNewPlotAndUpdateServer(symbols)

	go func() {
		for {
			select {
			case <-ticker.C:
				generateNewPlotAndUpdateServer(symbols)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	server.New()
	close(quit)
}

func generateNewPlotAndUpdateServer(symbols []string) {
	fmt.Println("INFO: Updating")

	to := time.Now()
	from := to.Add(-time.Duration((period + lookBackInterval) * time.Hour * 24))

	fmt.Println("INFO: Generating updated plot")
	plotWriter, err := plot.New(symbols, from, to, lookBackInterval)
	fmt.Println("INFO: Done generating updated plot")

	fmt.Println("INFO: Updating server with new plot")
	server.UpdateWriter(plotWriter, err)
	fmt.Println("INFO: Done updating server with new plot")

	fmt.Println("INFO: Done updating\nWaiting for next update...")
}
