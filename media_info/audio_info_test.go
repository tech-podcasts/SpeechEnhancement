package media_info

import (
	"reflect"
	"testing"
)

func TestAudioInfo(t *testing.T) {
	t.Run("AudioInfoMP3", func(t *testing.T) {
		target := AudioInfoRes{
			AudioStream: []AudioInfoAudioStream{
				{
					SampleRate:    "44100",
					Channels:      1,
					BitsPerSample: 0,
					BitRate:       "128000",
				},
			},
			Format: AudioInfoFormat{
				FormatName: "mp3",
				Duration:   "5.041633",
				Size:       "81083",
			},
		}
		got, err := AudioInfo("../tests/test.mp3")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, target) {
			t.Error("got: ", got, "target: ", target)
		}

	})
	t.Run("AudioInfoWAV", func(t *testing.T) {
		target := AudioInfoRes{
			AudioStream: []AudioInfoAudioStream{
				{
					SampleRate:    "44100",
					Channels:      1,
					BitsPerSample: 16,
					BitRate:       "705600",
				},
			},
			Format: AudioInfoFormat{
				FormatName: "wav",
				Duration:   "5.000000",
				Size:       "441044",
			},
		}
		got, err := AudioInfo("../tests/test.wav")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, target) {
			t.Error("got: ", got, "target: ", target)
		}
	})
}
