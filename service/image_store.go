package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
)

type ImageStore interface {
	Save(laptopID string, imageType string, imageData bytes.Buffer) (id string, err error)
}

type DiskImageStore struct {
	imageFolder string
	images      map[string]*ImageInfo
	mutex       sync.RWMutex
}

type ImageInfo struct {
	LaptopID string
	Type     string
	Path     string
}

func NewDiskImageStore(folder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: folder,
		images:      make(map[string]*ImageInfo),
	}
}

func (s *DiskImageStore) Save(laptopID string, imageType string, imageData bytes.Buffer) (id string, err error) {
	imageID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("cant not create uuid %w", err)
	}

	imagePath := fmt.Sprintf("./%s/%s%s", s.imageFolder, imageID.String(), imageType)
	file, err := os.Create(imagePath)
	if err != nil {
		_, err = os.Stat("./img")
		if os.IsNotExist(err) {
			os.MkdirAll("../../img", 0755)
		}
		file, err = os.Create(imagePath)
		if err != nil {
			return "", fmt.Errorf("cant not create file %w", err)
		}
	}

	writenBytes, err := imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("error when writing image file %v bytes %w", writenBytes, err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.images[imageID.String()] = &ImageInfo{
		LaptopID: laptopID,
		Type:     imageType,
		Path:     imagePath,
	}

	return imageID.String(), nil
}
