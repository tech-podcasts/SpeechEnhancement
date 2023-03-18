package process

import (
	"SpeechEnhancement/media_info"
	"SpeechEnhancement/meta"
	"log"
	"path/filepath"
	"testing"
)

type T meta.M

func (m T) Read(uuid string) (metadata meta.Metadata, err error) {
	return meta.DecodeRead(filePath(uuid))
}
func (m T) Write(uuid string, metadata meta.Metadata) error {
	return meta.EncodeWrite(filePath(uuid), metadata)
}
func filePath(uuid string) string {
	return filepath.Join("..", "tests", uuid, "meta.json")
}

func NewTest(uuid string, source string) (*Process, error) {
	dir := filepath.Join("..", "tests", uuid)
	sourcePath := filepath.Join(dir, source)
	mediaInfo, err := media_info.AudioInfo(sourcePath)
	var m T
	if err != nil {
		log.Printf("Error:NewProcess %e\n", err)
		err = meta.Set(m, uuid, meta.KeyStatus, meta.StatusDefault)
		log.Printf("Error:NewProcess %e\n", err)
		return &Process{}, err
	}
	return &Process{
		UUID:      uuid,
		Source:    sourcePath,
		MediaInfo: mediaInfo,
		TargetDir: dir,
		Context: Context{
			index: 0,
			Keys:  map[string]any{},
			meta:  m,
		},
	}, nil
}

func Test_Process(t *testing.T) {
	t.Run("ProcessChains", func(t *testing.T) {
		p, err := NewTest("7a8afb5d-e059-459c-a367-95f41436e326", "test1.mp3")
		if err != nil {
			t.Error(err)
		}
		p.SetChain().Handle()
	})
}
