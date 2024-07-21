package models

type ListImagesResponseImage struct {
	ID   string
	Name string
	Size int
}

type ListImagesResponse struct {
	Images []ListImagesResponseImage
}
