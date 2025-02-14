package manager

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pajtaand/dmap-zero/internal/common/database"
	errs "github.com/pajtaand/dmap-zero/internal/common/errors"
	"github.com/pajtaand/dmap-zero/internal/common/wrapper"
	"github.com/rs/zerolog/log"
)

type Image struct {
	id        string
	name      string
	reference string
	filename  string
	size      int

	mu       sync.RWMutex
	database *database.KVStore // TODO use interface and accept whatever
}

func NewImage(id, name, reference string, data []byte, database *database.KVStore) (*Image, error) {
	if database == nil {
		return nil, errors.New("database must not be nil")
	}

	fileName := fmt.Sprintf("img_%s", uuid.New().String())
	database.Set(fileName, data)

	return &Image{
		id:        id,
		name:      name,
		reference: reference,
		filename:  fileName,
		size:      len(data),
		database:  database,
	}, nil
}

func (i *Image) GetID() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.id
}

func (i *Image) GetName() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.name
}

func (i *Image) GetReference() string {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.reference
}

func (i *Image) GetData() ([]byte, error) {
	i.mu.RLock()
	fileName := i.filename
	i.mu.RUnlock()

	val, ok := i.database.Get(fileName)
	if !ok {
		return nil, fmt.Errorf("no file in database: fileName=%s", fileName)
	}

	data, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to parse image to bytes: fileName=%s", fileName)
	}
	return data, nil
}

func (i *Image) GetSize() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.size
}

func (i *Image) Cleanup() {
	i.mu.RLock()
	fileName := i.filename
	i.mu.RUnlock()

	i.database.Delete(fileName)
}

type ImageManager struct {
	mu            sync.RWMutex
	images        map[string]*Image
	dockerWrapper *wrapper.DockerClientWrapper
	database      *database.KVStore // TODO use interface and accept whatever
}

func NewImageManager(dockerWrapper *wrapper.DockerClientWrapper, database *database.KVStore) (*ImageManager, error) {
	log.Debug().Msg("Creating new ImageManager")

	if dockerWrapper == nil {
		return nil, errors.New("DockerClientWrapper must not be nil")
	}
	if database == nil {
		return nil, errors.New("database must not be nil")
	}

	return &ImageManager{
		images:        map[string]*Image{},
		dockerWrapper: dockerWrapper,
		database:      database,
	}, nil
}

func (mgr *ImageManager) AddImage(name string, data []byte) (*Image, error) {
	log.Info().Msgf("Adding new image: %s", name)

	reference, err := mgr.dockerWrapper.LoadImage(context.Background(), data)
	if err != nil {
		return nil, fmt.Errorf("failed to load image to docker: %v", err)
	}

	imageID := uuid.New().String()
	image, err := NewImage(imageID, name, reference, data, mgr.database)
	if err != nil {
		return nil, fmt.Errorf("failed to add new image: %v", err)
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.images[imageID] = image
	return image, err
}

func (mgr *ImageManager) AddImageWithID(id, name string, data []byte) (*Image, error) {
	log.Info().Msgf("Adding new image with id: %s (%s)", name, id)

	if mgr.ImageExists(id) {
		return nil, errs.ErrConflict
	}

	reference, err := mgr.dockerWrapper.LoadImage(context.Background(), data)
	if err != nil {
		return nil, fmt.Errorf("failed to load image to docker: %v", err)
	}

	image, err := NewImage(id, name, reference, data, mgr.database)
	if err != nil {
		return nil, fmt.Errorf("failed to add new image: %v", err)
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.images[id] = image
	return image, nil
}

func (mgr *ImageManager) GetImage(imageID string) (*Image, error) {
	log.Info().Msgf("Getting image: %s", imageID)

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	image, ok := mgr.images[imageID]
	if !ok {
		return nil, errs.ErrNotFound
	}

	return image, nil
}

func (mgr *ImageManager) ListImages() []*Image {
	log.Info().Msg("Listing all images")

	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	images := []*Image{}
	for _, image := range mgr.images {
		images = append(images, image)
	}
	return images
}

func (mgr *ImageManager) RemoveImage(imageID string) error {
	log.Info().Msgf("Removing image: %s", imageID)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	image, ok := mgr.images[imageID]
	if !ok {
		return errs.ErrNotFound
	}

	if err := mgr.dockerWrapper.RemoveImage(context.Background(), image.GetReference()); err != nil {
		return fmt.Errorf("faile dto remove docker image: %v", err)
	}

	image.Cleanup()

	delete(mgr.images, imageID)
	return nil
}

func (mgr *ImageManager) ImageExists(imageID string) bool {
	log.Info().Msgf("Checking if image exists: %s", imageID)
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	_, ok := mgr.images[imageID]
	return ok
}
