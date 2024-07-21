package handler

import (
	"errors"
	"fmt"
	"net/http"

	errs "github.com/andreepyro/dmap-zero/internal/common/errors"
	"github.com/andreepyro/dmap-zero/internal/common/utils"
	"github.com/andreepyro/dmap-zero/internal/controller/dto"
	"github.com/andreepyro/dmap-zero/internal/controller/rest/models"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

const (
	maxImageSize = 5 << 30 // 5*2^30 = 5GiB
)

type imageHandler struct {
	service ImageService
}

func NewImageHandler(service ImageService) *imageHandler {
	return &imageHandler{
		service: service,
	}
}

func (h *imageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxImageSize)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}
	defer file.Close()

	image, err := h.service.UploadImage(r.Context(), &dto.UploadImageRequest{
		Name: r.FormValue("name"),
		Src:  file,
	})
	if err != nil {
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, models.UploadImageResponse{
		ID: image.ID,
	})
}

func (h *imageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	imageID := chi.URLParam(r, "imageID")
	if imageID == "" {
		log.Info().Msg("imageID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	image, err := h.service.GetImage(r.Context(), &dto.GetImageRequest{
		ID: imageID,
	})
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("image with id '%s' doesn't exists", imageID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, models.GetImageResponse{
		Name: image.Name,
		Size: image.Size,
	})
}

func (h *imageHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	images, err := h.service.ListImages(r.Context(), &dto.ListImagesRequest{})
	if err != nil {
		panic(err)
	}

	imageList := []models.ListImagesResponseImage{}
	for _, image := range images.Images {
		imageList = append(imageList, models.ListImagesResponseImage{
			ID:   image.ID,
			Name: image.Name,
			Size: image.Size,
		})
	}
	utils.WriteResponse(w, http.StatusOK, models.ListImagesResponse{
		Images: imageList,
	})
}

func (h *imageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	log := zerolog.Ctx(r.Context())

	imageID := chi.URLParam(r, "imageID")
	if imageID == "" {
		log.Info().Msg("imageID is empty")
		utils.WriteErrorResponse(w, http.StatusBadRequest, nil)
		return
	}

	if _, err := h.service.DeleteImage(r.Context(), &dto.DeleteImageRequest{
		ID: imageID,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			log.Error().Err(err).Msg("")
			utils.WriteErrorResponse(w, http.StatusNotFound, fmt.Errorf("image with id '%s' doesn't exists", imageID))
			return
		}
		panic(err)
	}

	utils.WriteResponse(w, http.StatusOK, nil)
}
