package dto

import "io"

type UploadImageRequest struct {
	Name string
	Src  io.Reader
}

type UploadImageResponse struct {
	ID string
}

type GetImageRequest struct {
	ID string
}

type GetImageResponse struct {
	Name string
	Size int
}

type ListImagesRequest struct {
}

type ListImagesResponseImage struct {
	ID   string
	Name string
	Size int
}

type ListImagesResponse struct {
	Images []*ListImagesResponseImage
}

type DeleteImageRequest struct {
	ID string
}

type DeleteImageResponse struct {
}
