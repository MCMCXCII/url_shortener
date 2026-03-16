package repository

import (
	"encoding/json"
	"os"
)

type FileRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	file     *os.File
	encoder  *json.Encoder
	filename string
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileStorage{
		file:     file,
		encoder:  json.NewEncoder(file),
		filename: filename,
	}, nil
}

func (f *FileStorage) WriteToFile(record FileRecord) error {
	return f.encoder.Encode(record)
}

func LoadFromFile(filename string) ([]FileRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []FileRecord
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var rec FileRecord
		if err := decoder.Decode(&rec); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}

	return records, nil
}
