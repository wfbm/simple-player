package player

import (
	"github.com/spf13/cobra"
)

func Play() *cobra.Command {
	return &cobra.Command{
		Use:   "play",
		Short: "play some song",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			StartPlaying(args[0])
		},
	}
}

func StartPlaying(song string) {

	err := Start(song)

	if err != nil {
		panic(err)
	}
}
