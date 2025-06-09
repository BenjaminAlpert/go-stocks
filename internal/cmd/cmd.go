package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	periodFlagDefault   = 20 // 20 years
	periodFlagName      = "period"
	periodFlagShorthand = "p"
	periodFlagUsage     = "number of years to display"

	intervalFlagDefault  = 365 // 365 days
	intervalFlagName     = "interval"
	intervalFlagShothand = "i"
	internalFlagUsage    = "number of days to look back prior to and average over"

	updateFrequencyFlagDefault   = 4 // 4 hours
	updateFrequencyFlagName      = "update-frequency"
	updateFrequencyFlagShorthand = "u"
	updateFrequencyFlagUsage     = "number hours between updates"

	symbolsFlagName      = "symbol"
	symbolsFlagShorthand = "s"
	symbolsFlagUsage     = "ticker symbol(s)"
)

var (
	symbolsFlagDefault = []string{"dia", "spy", "vt"}
)

type Callback func(period int, interval int, updateFrequency int, symbols []string)

func New(callback Callback) error {
	rootCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			period, err := cmd.Flags().GetInt(periodFlagName)
			if err != nil {
				return fmt.Errorf("error getting %s flag: %s", periodFlagName, err)
			}
			interval, err := cmd.Flags().GetInt(intervalFlagName)
			if err != nil {
				return fmt.Errorf("error getting %s flag: %s", intervalFlagName, err)
			}

			updateFrequency, err := cmd.Flags().GetInt(updateFrequencyFlagName)
			if err != nil {
				return fmt.Errorf("error getting %s flag: %s", updateFrequencyFlagName, err)
			}

			symbols, err := cmd.Flags().GetStringArray(symbolsFlagName)
			if err != nil {
				return fmt.Errorf("error getting %s flag: %s", symbolsFlagName, err)
			}

			callback(period, interval, updateFrequency, symbols)
			return nil
		},
	}
	rootCmd.Flags().IntP(periodFlagName, periodFlagShorthand, periodFlagDefault, periodFlagUsage)
	rootCmd.Flags().IntP(intervalFlagName, intervalFlagShothand, intervalFlagDefault, internalFlagUsage)
	rootCmd.Flags().IntP(updateFrequencyFlagName, updateFrequencyFlagShorthand, updateFrequencyFlagDefault, updateFrequencyFlagUsage)
	rootCmd.Flags().StringArrayP(symbolsFlagName, symbolsFlagShorthand, symbolsFlagDefault, symbolsFlagUsage)

	return rootCmd.Execute()
}
