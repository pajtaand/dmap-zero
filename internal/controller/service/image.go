package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/pajtaand/dmap-zero/internal/common/constants"
	errs "github.com/pajtaand/dmap-zero/internal/common/errors"
	"github.com/pajtaand/dmap-zero/internal/controller/dto"
	"github.com/pajtaand/dmap-zero/internal/controller/manager"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
)

type imageService struct {
	imageManager  *manager.ImageManager
	agentManager  *manager.AgentManager
	moduleManager *manager.ModuleManager
}

func NewImageService(imageManager *manager.ImageManager, agentManager *manager.AgentManager, moduleManager *manager.ModuleManager) (*imageService, error) {
	if imageManager == nil {
		return nil, errors.New("ImageManager must not be nil")
	}
	if agentManager == nil {
		return nil, errors.New("AgentManager must not be nil")
	}
	if moduleManager == nil {
		return nil, errors.New("ModuleManager must not be nil")
	}

	return &imageService{
		imageManager:  imageManager,
		agentManager:  agentManager,
		moduleManager: moduleManager,
	}, nil
}

func (svc *imageService) UploadImage(ctx context.Context, request *dto.UploadImageRequest) (*dto.UploadImageResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Upload image request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	if request.Src == nil {
		return nil, errors.New("src must not be nil")
	}

	data, err := io.ReadAll(request.Src)
	if err != nil {
		return nil, err
	}

	image, err := svc.imageManager.AddImage(request.Name, data)
	if err != nil {
		return nil, err
	}
	imageID := image.GetID()

	for _, agent := range svc.agentManager.ListAgents() {
		go func(ctx context.Context, agent *manager.Agent) {
			agentID := agent.GetID()

			c := agent.GetImageServiceClient()
			if c == nil {
				return
			}
			log.Info().Msgf("Pushing image to agent: agentID=%s, imageID=%s", agentID, imageID)

			stream, err := c.PushImage(ctx)
			if err != nil {
				log.Info().Msgf("Failed to create stream to agentID=%s: %v", agentID, err)
				return
			}

			for start := 0; start < len(data); start += constants.AgentImageStreamChunkSize {
				end := start + constants.AgentImageStreamChunkSize
				if end > len(data) {
					end = len(data)
				}
				if err := stream.Send(&pb.ImageStreamData{
					Id:      imageID,
					Name:    request.Name,
					Content: data[start:end],
				}); err != nil {
					log.Info().Msgf("Failed to stream image to agentID=%s: %v", agentID, err)
					return
				}
			}

			if _, err := stream.CloseAndRecv(); err != nil {
				log.Info().Msgf("Failed to receive response from agentID=%s: %v", agentID, err)
				return
			}
			log.Info().Msgf("Image push finished: agentID=%s, imageID=%s", agentID, imageID)
		}(context.Background(), agent)
	}

	return &dto.UploadImageResponse{
		ID: imageID,
	}, nil

}

func (svc *imageService) GetImage(ctx context.Context, request *dto.GetImageRequest) (*dto.GetImageResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Get image request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	image, err := svc.imageManager.GetImage(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %v", err)
	}

	return &dto.GetImageResponse{
		Name: image.GetName(),
		Size: image.GetSize(),
	}, nil
}

func (svc *imageService) ListImages(ctx context.Context, request *dto.ListImagesRequest) (*dto.ListImagesResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("List images request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	images := make([]*dto.ListImagesResponseImage, 0)
	for _, image := range svc.imageManager.ListImages() {
		images = append(images, &dto.ListImagesResponseImage{
			ID:   image.GetID(),
			Name: image.GetName(),
			Size: image.GetSize(),
		})
	}
	return &dto.ListImagesResponse{
		Images: images,
	}, nil
}

func (svc *imageService) DeleteImage(ctx context.Context, request *dto.DeleteImageRequest) (*dto.DeleteImageResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Delete image request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	image, err := svc.imageManager.GetImage(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %v", err)
	}

	for _, module := range svc.moduleManager.ListModules() {
		if module.GetImage() == image.GetID() {
			return nil, errs.ErrNotAllowed
		}
	}

	for _, agent := range svc.agentManager.ListAgents() {
		c := agent.GetImageServiceClient()
		if c == nil {
			continue
		}
		log.Info().Msgf("Sending image removal request: agentID=%s, imageID=%s", agent.GetID(), image.GetID())

		if _, err := c.RemoveImage(ctx, &pb.ImageIdentifier{
			Id: image.GetID(),
		}); err != nil {
			log.Info().Msgf("Failed to create stream to agentID=%s: %v", agent.GetID(), err)
			continue
		}
		log.Info().Msgf("Image removal resposne: imageID=%s, agentID=%s", image.GetID(), agent.GetID())
	}

	if err := svc.imageManager.RemoveImage(request.ID); err != nil {
		return nil, fmt.Errorf("failed to remove image: %v", err)
	}

	return &dto.DeleteImageResponse{}, nil
}
