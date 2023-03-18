package media_info

import (
	"encoding/json"
	"log"
	"os/exec"
)

type AudioInfoAudioStream struct {
	SampleRate    string `json:"sample_rate"`
	Channels      int    `json:"channels"`
	BitsPerSample int    `json:"bits_per_sample"`
	BitRate       string `json:"bit_rate"`
}

type AudioInfoFormat struct {
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
}

type AudioInfoRes struct {
	AudioStream []AudioInfoAudioStream `json:"streams"`
	Format      AudioInfoFormat        `json:"format"`
}

func AudioInfo(file string) (AudioInfoRes, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-hide_banner", "-select_streams", "a", "-show_streams", file)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("Error:AudioInfo Command %s\n", out)
		return AudioInfoRes{}, err
	}
	var res AudioInfoRes
	err = json.Unmarshal(out, &res)
	if err != nil {
		return AudioInfoRes{}, err
	}
	return res, nil
}
