package main

import (
	"fmt"
	"time"

	"github.com/BenjaminAlpert/go-stocks/internal/cmd"
	"github.com/BenjaminAlpert/go-stocks/internal/plot"
	"github.com/BenjaminAlpert/go-stocks/internal/server"
)

func main() {
	err := cmd.New(func(period int, interval int, updateFrequency int, symbols []string) {
		ticker := time.NewTicker(time.Duration(updateFrequency) * time.Hour)

		go func() {
			for {
				fmt.Println("[INFO] Updating")

				to := time.Now()
				from := to.Add(-time.Duration((period*365+interval)*24) * time.Hour)

				fmt.Println("[INFO] Generating updated plot")
				plotWriter, err := plot.New(symbols, from, to, interval)
				if err != nil {
					fmt.Printf("[ERROR] failed to generate a updated plot, %s\n", err)
					<-ticker.C // Wait until ticker to update
					continue
				}
				fmt.Println("[INFO] Done generating updated plot")

				fmt.Println("[INFO] Updating server with new plot")
				server.UpdateWriter(plotWriter, err)
				fmt.Println("[INFO] Done updating server with new plot")

				fmt.Println("[INFO] Done updating")
				fmt.Println("[INFO] Waiting for next update...")

				<-ticker.C // Wait until ticker to update
			}
		}()

		server.New()
		ticker.Stop()
	})

	if err != nil {
		fmt.Printf("[ERROR] failed to parse command: %s\n", err)
	}
}
