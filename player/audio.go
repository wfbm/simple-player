package player

import (
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

func Start(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	streamer, format, err := wav.Decode(file)
	if err != nil {
		return err
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	playback := buffer.Streamer(0, buffer.Len())
	spectrum := buffer.Streamer(0, buffer.Len())

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(playback, beep.Callback(func() { done <- true })))

	totalSeconds := streamer.Len() / int(format.SampleRate)

	go func() {
		for {
			spectrum.Seek(playback.Position())
			show(totalSeconds, spectrum, format)
			time.Sleep(time.Second / 10)
		}
	}()

	<-done
	clearSpectrum()
	displayPlayingBar(0, totalSeconds)

	return nil
}
