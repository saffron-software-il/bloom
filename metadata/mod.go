package metadata

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

type Metadata struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	Build     int    `json:"build"`
	Author    string `json:"author"`
	Icon      string `json:"icon"`
	Generator string `json:"generator"`
}

func NewMetadata(id, name, version string, build int, author, generator, icon string) Metadata {
	return Metadata{
		Id:        id,
		Name:      name,
		Version:   version,
		Build:     build,
		Author:    author,
		Icon:      icon,
		Generator: generator,
	}
}

func (m *Metadata) Save(dest string) (error, bool) {
	data, err := json.Marshal(m)
	if err != nil {
		return err, false
	}

	filePath := filepath.Join(dest, "index.json")

	file, err := os.Create(filePath)
	if err != nil {
		return err, false
	}

	w := bufio.NewWriter(file)
	_, err = w.Write(data)
	if err != nil {
		return err, false
	}

	err = w.Flush()
	if err != nil {
		return err, false
	}

	return nil, true
}
