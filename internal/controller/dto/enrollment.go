package dto

import "time"

type CreateEnrollmentRequest struct {
	ID string
}

type CreateEnrollmentResponse struct {
	JWT       string
	ExpiresAt time.Time
}

type GetEnrollmentRequest struct {
	ID string
}

type GetEnrollmentResponse struct {
	JWT       string
	ExpiresAt time.Time
}

type DeleteEnrollmentRequest struct {
	ID string
}

type DeleteEnrollmentResponse struct {
}
