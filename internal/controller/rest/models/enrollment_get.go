package models

import "time"

type GetEnrollmentResponse struct {
	JWT       string
	ExpiresAt time.Time
}
