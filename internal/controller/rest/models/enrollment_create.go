package models

import "time"

type CreateEnrollmentResponse struct {
	JWT       string
	ExpiresAt time.Time
}
