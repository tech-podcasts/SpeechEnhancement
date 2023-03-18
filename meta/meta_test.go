package meta

import (
	"path/filepath"
	"reflect"
	"testing"
)

func testFilePath(uuid string) string {
	return filepath.Join("..", "tests", uuid, "meta.json")
}

func Test_write(t *testing.T) {
	t.Run("MetaWrite", func(t *testing.T) {
		metadata := Metadata{
			Source: "test.mp3",
			Target: "test.wav",
			Status: 3,
		}
		if err := EncodeWrite(testFilePath("7a8afb5d-e059-459c-a367-95f41436e326"), metadata); err != nil {
			t.Error(err)
		}
	})
}

func Test_read(t *testing.T) {
	t.Run("MetaRead", func(t *testing.T) {
		target := Metadata{
			Source: "test.mp3",
			Target: "test.wav",
			Status: 3,
		}
		if got, err := DecodeRead(testFilePath("7a8afb5d-e059-459c-a367-95f41436e326")); err == nil {
			if !reflect.DeepEqual(got, target) {
				t.Error("got: ", got, "target: ", target)
			}
		} else {
			t.Error(err)
		}
	})
}
