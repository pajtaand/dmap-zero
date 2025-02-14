package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/pajtaand/dmap-zero/internal/agent/manager"
	pb "github.com/pajtaand/dmap-zero/internal/proto"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/emptypb"
)

type imageService struct {
	pb.UnimplementedImageServiceServer

	imageManager *manager.ImageManager
}

func NewImageService(imageManager *manager.ImageManager) (pb.ImageServiceServer, error) {
	if imageManager == nil {
		return nil, errors.New("ImageManager must not be nil")
	}

	return &imageService{
		imageManager: imageManager,
	}, nil
}

func (svc *imageService) CheckImage(ctx context.Context, identifier *pb.ImageIdentifier) (*pb.ResourceExistResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msgf("Check image request: imageID=%s", identifier.Id)

	isPresent := svc.imageManager.ImageExists(identifier.Id)
	log.Info().Msgf("Image exists response: imageID=%s, isPresent=%t", identifier.Id, isPresent)

	return &pb.ResourceExistResponse{
		IsPresent: isPresent,
	}, nil
}

func (svc *imageService) GetImage(ctx context.Context, identifier *pb.ImageIdentifier) (*pb.ImageInfo, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msgf("Get image request: imageID=%s", identifier.Id)

	if !svc.imageManager.ImageExists(identifier.Id) {
		err := errors.New("image doesn't exist")
		log.Error().Err(err).Msg("")
		return nil, err
	}

	image, err := svc.imageManager.GetImage(identifier.Id)
	if err != nil {
		err := fmt.Errorf("failed to get image: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	log.Info().Msgf("Retrieving image information: imageID=%s", identifier.Id)
	return &pb.ImageInfo{
		Id:   image.GetID(),
		Name: image.GetName(),
		Size: int64(image.GetSize()),
	}, nil
}

func (svc *imageService) PushImage(stream pb.ImageService_PushImageServer) error {
	log := zerolog.Ctx(stream.Context())
	log.Info().Msg("Push image request")

	var imageID, imageName string
	var imageData []byte
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Info().Msgf("Image successfully received: imageID=%s, imageName=%s", imageID, imageName)

				if _, err := svc.imageManager.AddImageWithID(imageID, imageName, imageData); err != nil {
					err := fmt.Errorf("failed to add image to the ImageManager: imageID=%s, imageName=%s", imageID, imageName)
					log.Error().Err(err).Msg("")
					return err
				}

				return stream.SendAndClose(&emptypb.Empty{})
			}
			log.Info().Msgf("Image upload failed: imageID=%s, imageName=%s, err: %v", imageID, imageName, err)
			return err
		}
		if imageName == "" {
			imageID = chunk.Id
			imageName = chunk.Name
			log.Info().Msgf("Image upload started: imageID=%s, imageName=%s", imageID, imageName)
		}
		imageData = append(imageData, chunk.Content...)
	}
}

func (svc *imageService) RemoveImage(ctx context.Context, identifier *pb.ImageIdentifier) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msgf("Remove image request: imageID=%s", identifier.Id)

	if !svc.imageManager.ImageExists(identifier.Id) {
		err := errors.New("image doesn't exist")
		log.Error().Err(err).Msg("")
		return nil, err
	}

	if err := svc.imageManager.RemoveImage(identifier.Id); err != nil {
		err := fmt.Errorf("failed to remove image: %v", err)
		log.Error().Err(err).Msg("")
		return nil, err
	}

	log.Info().Msgf("Image removed: imageID=%s", identifier.Id)

	return &emptypb.Empty{}, nil
}
