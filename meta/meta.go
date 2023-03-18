package meta

import (
	"SpeechEnhancement/upload"
	"encoding/json"
	"os"
	"path/filepath"
)

type Metadata struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	Status     int    `json:"status"`
	CreateTime int    `json:"createTime"`
}

const (
	StatusDefault = iota
	StatusTranscode
	StatusEnhance
	StatusNormalize
	StatusComplete
	StatusError

	KeySource     = "source"
	KeyTarget     = "target"
	KeyStatus     = "status"
	KeyCreateTime = "createTime"
)

type Meta interface {
	Read(uuid string) (metadata Metadata, err error)
	Write(uuid string, metadata Metadata) error
}

func DecodeRead(filePath string) (metadata Metadata, err error) {
	var file *os.File
	file, err = os.Open(filePath)
	defer file.Close()
	if err != nil {
		return metadata, err
	}
	err = json.NewDecoder(file).Decode(&metadata)
	if err != nil {
		return metadata, err
	}
	return
}

func EncodeWrite(filePath string, metadata Metadata) error {
	file, err := os.Create(filePath)
	defer file.Close()
	if err != nil {
		return err
	}
	err = json.NewEncoder(file).Encode(metadata)
	if err != nil {
		return err
	}
	return nil
}

type M struct{}

func (m M) Read(uuid string) (metadata Metadata, err error) {
	return DecodeRead(filePath(uuid))
}
func (m M) Write(uuid string, metadata Metadata) error {
	return EncodeWrite(filePath(uuid), metadata)
}

func Set(m Meta, uuid string, key string, value any) error {
	metadata, err := m.Read(uuid)
	if err != nil {
		return err
	}
	switch key {
	case KeySource:
		metadata.Source = value.(string)
	case KeyTarget:
		metadata.Target = value.(string)
	case KeyStatus:
		metadata.Status = value.(int)
	case KeyCreateTime:
		metadata.CreateTime = value.(int)
	}
	return m.Write(uuid, metadata)
}

func Get(m Meta, uuid string, key string) (value any, err error) {
	metadata, err := m.Read(uuid)
	if err != nil {
		return nil, err
	}
	switch key {
	case KeySource:
		return metadata.Source, nil
	case KeyTarget:
		return metadata.Target, nil
	case KeyStatus:
		return metadata.Status, nil
	case KeyCreateTime:
		return metadata.CreateTime, nil
	}
	return nil, nil
}

func filePath(uuid string) string {
	return filepath.Join(upload.DefaultUploadPath, uuid, "meta.json")
}
