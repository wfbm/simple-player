package player

import (
	"fmt"
	"math/cmplx"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/mjibson/go-dsp/fft"
	//"github.com/mjibson/go-dsp/fft"
)

const (
	width  = 60
	height = 20
)

type Spectrum struct {
	samples            [][2]float64
	sampleAmount       int
	format             beep.Format
	streamer           beep.StreamSeeker
	totalTimeInSeconds int
}

func DisplaySpectrum(streamer beep.StreamSeeker, format beep.Format) beep.Streamer {
	//ticker := time.NewTicker(time.Second / 10)
	spectrumChan := make(chan Spectrum, 1)

	totalSeconds := streamer.Len() / int(format.SampleRate)

	//	go show(ticker, spectrumChan)

	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for len(samples) > 0 {
			sn, sok := streamer.Stream(samples)
			n, ok = n+sn, ok || sok
			if !sok {
				break
			}

			spectrumChan <- Spectrum{
				samples:            samples,
				sampleAmount:       sn,
				format:             format,
				streamer:           streamer,
				totalTimeInSeconds: totalSeconds,
			}

			samples = samples[sn:]
		}

		return n, ok
	})
}

func show(totalTimeInSeconds int, streamer beep.StreamSeeker, format beep.Format) {

	samples := make([][2]float64, format.SampleRate.N(time.Second/10))
	n, ok := streamer.Stream(samples)

	if !ok {
		return
	}

	combinedChannels := make([]float64, n)
	for i := 0; i < n; i++ {
		combinedChannels[i] = (samples[i][0] + samples[i][1]) / 2
	}

	fftResult := fft.FFTReal(combinedChannels)
	curSecond := streamer.Position() / int(format.SampleRate)

	displaySpectrum(fftResult)
	displayPlayingBar(curSecond, totalTimeInSeconds)
}

func displaySpectrum(fftResult []complex128) {
	magnitudes := make([]float64, len(fftResult)/2)
	for i := range magnitudes {
		magnitudes[i] = cmplx.Abs(fftResult[i])
	}

	scaledMagnitudes := scaleMagnitudes(magnitudes, height)

	var rows []string
	for i := 0; i < height; i++ {
		row := ""
		for j := 0; j < width; j++ {
			if i < height-int(scaledMagnitudes[j]) {
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

	barWidth := width
	curPos := (barWidth * position) / total

	for i := 0; i < barWidth; i++ {

		if i == curPos {
			playingBar += "█"
		}

		playingBar += "▬"
	}

	p := time.Second * time.Duration(total)
	playingBar += fmt.Sprintf("%v", p)

	fmt.Println("")
	fmt.Println(style.Render(playingBar))
	fmt.Println("")

}

func clearSpectrum() {

	var rows []string
	for i := 0; i < height; i++ {
		row := ""
		for j := 0; j < width; j++ {

			row += " "
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
