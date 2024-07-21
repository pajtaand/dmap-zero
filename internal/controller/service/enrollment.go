package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andreepyro/dmap-zero/internal/common/constants"
	errs "github.com/andreepyro/dmap-zero/internal/common/errors"
	"github.com/andreepyro/dmap-zero/internal/common/wrapper"
	"github.com/andreepyro/dmap-zero/internal/controller/dto"
	"github.com/andreepyro/dmap-zero/internal/controller/manager"
	"github.com/go-openapi/strfmt"
	"github.com/rs/zerolog"
)

type enrollmentService struct {
	agentManager    *manager.AgentManager
	openZitiWrapper *wrapper.OpenZitiManagementWrapper
}

func NewEnrollmentService(agentManager *manager.AgentManager, openZitiWrapper *wrapper.OpenZitiManagementWrapper) (*enrollmentService, error) {
	if agentManager == nil {
		return nil, errors.New("AgentManager must not be nil")
	}

	if openZitiWrapper == nil {
		return nil, errors.New("OpenZitiManagementWrapper must not be nil")
	}

	return &enrollmentService{
		agentManager:    agentManager,
		openZitiWrapper: openZitiWrapper,
	}, nil
}

func (svc *enrollmentService) CreateEnrollment(ctx context.Context, request *dto.CreateEnrollmentRequest) (*dto.CreateEnrollmentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Create enrollment request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	agent, err := svc.agentManager.GetAgent(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %v", err)
	}

	if identityID := agent.GetIdentityID(); identityID != "" {
		return nil, errors.New("identity already exist")
	}

	identityID, err := svc.openZitiWrapper.CreateIdentity(agent.GetID(), constants.OpenZitiAdminAgent, []string{constants.OpenZitiRoleAgent})
	if err != nil {
		return nil, fmt.Errorf("failed to create identity for agent: %v", err)
	}
	agent.SetIdentityID(identityID)

	ExpiresAt := time.Now().Add(constants.OpenZitiEnrollmentTokenValidity)
	enrollmentID, err := svc.openZitiWrapper.CreateEnrollment(agent.GetIdentityID(), strfmt.DateTime(ExpiresAt))
	if err != nil {
		return nil, fmt.Errorf("failed to create enrollment: %v", err)
	}

	JWT, err := svc.openZitiWrapper.GetEnrollmentToken(enrollmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enrollment token: %v", err)
	}

	return &dto.CreateEnrollmentResponse{
		JWT:       JWT,
		ExpiresAt: ExpiresAt,
	}, nil
}

func (svc *enrollmentService) GetEnrollment(ctx context.Context, request *dto.GetEnrollmentRequest) (*dto.GetEnrollmentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Get enrollment request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	agent, err := svc.agentManager.GetAgent(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %v", err)
	}

	identityID := agent.GetIdentityID()
	if identityID == "" {
		return nil, errs.ErrNotFound
	}

	detail, err := svc.openZitiWrapper.GetIdentityDetail(identityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity detail: %v", err)
	}

	return &dto.GetEnrollmentResponse{
		JWT:       detail.Enrollment.Ott.JWT,
		ExpiresAt: time.Time(detail.Enrollment.Ott.ExpiresAt),
	}, nil
}

func (svc *enrollmentService) DeleteEnrollment(ctx context.Context, request *dto.DeleteEnrollmentRequest) (*dto.DeleteEnrollmentResponse, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Msg("Delete enrollment request")

	if request == nil {
		return nil, errors.New("request must not be nil")
	}

	agent, err := svc.agentManager.GetAgent(request.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %v", err)
	}

	if identityID := agent.GetIdentityID(); identityID != "" {
		if err := svc.openZitiWrapper.DeleteIdentity(identityID); err != nil {
			return nil, fmt.Errorf("failed to remove identity for agent: %v", err)
		}
		agent.SetIdentityID("")
	}

	return &dto.DeleteEnrollmentResponse{}, nil
}
