package main

import (
	"fmt"
	"math/cmplx"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
	"github.com/mjibson/go-dsp/fft"
)

const (
	terminalWidth  = 60
	terminalHeight = 20
)

func main() {
	f, err := os.Open("song.wav")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		fmt.Println("Error decoding WAV file:", err)
		return
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	playbackStreamer := buffer.Streamer(0, buffer.Len())
	spectrumStreamer := buffer.Streamer(0, buffer.Len())
	done := make(chan bool)
	speaker.Play(playbackStreamer)

	go func() {
		samples := make([][2]float64, format.SampleRate.N(time.Second/10))
		for {
			n, ok := spectrumStreamer.Stream(samples)
			if !ok {
				break
			}

			combinedChannels := make([]float64, n)
			for i := 0; i < n; i++ {
				combinedChannels[i] = (samples[i][0] + samples[i][1]) / 2
			}

			fftResult := fft.FFTReal(combinedChannels)

			displaySpectrum(fftResult)
			curSec := playbackStreamer.Position() / int(format.SampleRate)
			totalSecs := playbackStreamer.Len() / int(format.SampleRate)

			displayPlayingBar(curSec, totalSecs)

			time.Sleep(time.Second / 10)
			if curSec == totalSecs {
				done <- true
				break
			}
		}
	}()

	<-done
}

func displaySpectrum(fftResult []complex128) {
	magnitudes := make([]float64, len(fftResult)/2)
	for i := range magnitudes {
		magnitudes[i] = cmplx.Abs(fftResult[i])
	}

	scaledMagnitudes := scaleMagnitudes(magnitudes, terminalHeight)

	var rows []string
	for i := 0; i < terminalHeight; i++ {
		row := ""
		for j := 0; j < terminalWidth; j++ {
			if i < terminalHeight-int(scaledMagnitudes[j]) {
				row += " "
			} else {
				row += "█"
			}
		}
		rows = append(rows, row)
	}

	fmt.Print("\033[H\033[2J")

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("0"))

	for _, row := range rows {
		fmt.Println(style.Render(row))
	}
}

func displayPlayingBar(position, total int) {

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("0"))

	playingBar := "▍▍ "

	barWidth := terminalWidth
	curPos := (barWidth / total) * position

	for i := 0; i < barWidth; i++ {

		if i == curPos {
			playingBar += "█"
		}

		playingBar += "▬"
	}
	fmt.Println("")
	fmt.Println(style.Render(playingBar))
	fmt.Println("")

}

func scaleMagnitudes(magnitudes []float64, maxHeight int) []float64 {
	maxMagnitude := 0.0
	for _, magnitude := range magnitudes {
		if magnitude > maxMagnitude {
			maxMagnitude = magnitude
		}
	}

	scaled := make([]float64, len(magnitudes))
	for i, magnitude := range magnitudes {
		scaled[i] = (magnitude / maxMagnitude) * float64(maxHeight)
	}

	return scaled
}
