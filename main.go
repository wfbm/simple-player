package main

import (
	"simple-player/player"

	"github.com/spf13/cobra"
)

const (
	terminalWidth  = 60
	terminalHeight = 20
)

func main() {

	rootCmd := cobra.Command{}
	rootCmd.AddCommand(player.Play())
	rootCmd.Execute()
}
