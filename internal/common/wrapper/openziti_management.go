package wrapper

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/andreepyro/dmap-zero/internal/common/constants"
	"github.com/go-openapi/strfmt"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router"
	"github.com/openziti/edge-api/rest_management_api_client/edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/enrollment"
	"github.com/openziti/edge-api/rest_management_api_client/identity"
	"github.com/openziti/edge-api/rest_management_api_client/service"
	"github.com/openziti/edge-api/rest_management_api_client/service_edge_router_policy"
	"github.com/openziti/edge-api/rest_management_api_client/service_policy"
	"github.com/openziti/edge-api/rest_model"
	edge_apis "github.com/openziti/sdk-golang/edge-apis"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/rs/zerolog/log"
)

const (
	enrollmentMethod  = "ott"
	identityType      = rest_model.IdentityTypeDevice
	identityListLimit = int64(999)
)

type OpenZitiManagementWrapper struct {
	cfg              *ziti.Config
	managementClient *edge_apis.ManagementApiClient
	apiSession       edge_apis.ApiSession
}

func NewOpenZitiManagementWrapper(cfg *ziti.Config) (*OpenZitiManagementWrapper, error) {
	log.Debug().Msg("Creating new OpenZitiManagementWrapper")

	if cfg == nil {
		return nil, errors.New("OpenZiti config is nil")
	}

	wrapper := &OpenZitiManagementWrapper{
		cfg: cfg,
	}

	if err := wrapper.authenticate(); err != nil {
		return nil, err
	}
	go wrapper.authenticationLoop()

	return wrapper, nil
}

func (w *OpenZitiManagementWrapper) authenticate() error {
	log.Debug().Msg("Authenticating to the OpenZiti controller")

	apiUrl, _ := url.Parse(w.cfg.ZtAPI)
	apiUrl.Path = edge_apis.ManagementApiPath
	credentials := w.cfg.Credentials
	managementClient := edge_apis.NewManagementApiClient([]*url.URL{apiUrl}, credentials.GetCaPool(), nil)

	var configTypes []string
	apiSession, err := managementClient.Authenticate(credentials, configTypes)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}
	w.managementClient = managementClient
	w.apiSession = apiSession
	return nil
}

func (w *OpenZitiManagementWrapper) authenticationLoop() {
	for {
		t := w.apiSession.GetExpiresAt()
		if t == nil {
			log.Error().Msg("OpenZiti session has no expire time set")
			return
		}

		log.Debug().Msgf("OpenZiti session expires in %v, waiting until then...", t)
		time.Sleep(time.Until(*t) - constants.OpenZitiManagementAPIReloadReserve)

		if err := w.authenticate(); err != nil {
			log.Error().Err(err).Msg("OpenZiti authentication failed")
			return
		}
	}
}

func (w *OpenZitiManagementWrapper) CreateIdentity(identityName string, isAdmin bool, roleAttributes rest_model.Attributes) (string, error) {
	t := identityType
	log.Debug().Msgf("Creating new OpenZiti identity: name=%s, type=%s, isAdmin=%t, roleAttributes=%+v", identityName, t, isAdmin, roleAttributes)
	newIdentity, err := w.managementClient.API.Identity.CreateIdentity(&identity.CreateIdentityParams{
		Identity: &rest_model.IdentityCreate{
			Name:           &identityName,
			IsAdmin:        &isAdmin,
			Type:           &t,
			RoleAttributes: &roleAttributes,
		},
	}, w.apiSession)
	if err != nil {
		return "", err
	}
	identityID := newIdentity.Payload.Data.ID
	log.Debug().Msgf("New identity created: identityID=%s", identityID)
	return identityID, nil
}

func (w *OpenZitiManagementWrapper) GetIdentityDetail(identityID string) (*rest_model.IdentityDetail, error) {
	log.Debug().Msgf("Getting OpenZiti identity details for: %s", identityID)
	resp, err := w.managementClient.API.Identity.DetailIdentity(&identity.DetailIdentityParams{
		ID: identityID,
	}, w.apiSession)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Received OpenZiti identity details for: %s", identityID)
	return resp.Payload.Data, nil
}

func (w *OpenZitiManagementWrapper) ListIdentityDetails() ([]*rest_model.IdentityDetail, error) {
	log.Debug().Msgf("Listing OpenZiti identity details")
	limit := identityListLimit
	resp, err := w.managementClient.API.Identity.ListIdentities(&identity.ListIdentitiesParams{
		Limit: &limit,
	}, w.apiSession)
	if err != nil {
		return nil, err
	}
	identityDetails := resp.Payload.Data
	log.Debug().Msgf("Got %d OpenZIti identity details", len(identityDetails))
	if len(identityDetails) == int(limit) {
		log.Warn().Msgf("Number of OpenZIti identity details is equal to the list limit. There might be more identities which are ignored now!")
	}
	return identityDetails, nil
}

func (w *OpenZitiManagementWrapper) DeleteIdentity(identityID string) error {
	log.Debug().Msgf("Deleting OpenZiti identity: %s", identityID)
	_, err := w.managementClient.API.Identity.DeleteIdentity(&identity.DeleteIdentityParams{
		ID: identityID,
	}, w.apiSession)
	if err != nil {
		return err
	}
	log.Debug().Msgf("OpenZiti identity deleted: %s", identityID)
	return nil
}

func (w *OpenZitiManagementWrapper) CreateEnrollment(identityID string, expiresAt strfmt.DateTime) (string, error) {
	m := enrollmentMethod
	log.Debug().Msgf("Creating OpenZiti enrollment token: identityID=%s, method=%s", identityID, m)
	newEnrollment, err := w.managementClient.API.Enrollment.CreateEnrollment(&enrollment.CreateEnrollmentParams{
		Enrollment: &rest_model.EnrollmentCreate{
			ExpiresAt:  &expiresAt,
			IdentityID: &identityID,
			Method:     &m,
		},
	}, w.apiSession)
	if err != nil {
		return "", err
	}
	enrollmentID := newEnrollment.Payload.Data.ID
	log.Debug().Msgf("OpenZiti enrollment token created with enrollmentID=%s", enrollmentID)
	return enrollmentID, nil
}

func (w *OpenZitiManagementWrapper) GetEnrollmentToken(enrollmentID string) (string, error) {
	log.Debug().Msgf("Getting OpenZiti enrollment token for enrollment: %s", enrollmentID)
	enrollmentDetail, err := w.managementClient.API.Enrollment.DetailEnrollment(&enrollment.DetailEnrollmentParams{
		ID: enrollmentID,
	}, w.apiSession)
	if err != nil {
		return "", err
	}
	return enrollmentDetail.Payload.Data.JWT, nil
}

func (w *OpenZitiManagementWrapper) DeleteEnrollment(enrollmentID string) error {
	log.Debug().Msgf("Deleting OpenZiti enrollment: %s", enrollmentID)
	_, err := w.managementClient.API.Enrollment.DeleteEnrollment(&enrollment.DeleteEnrollmentParams{
		ID: enrollmentID,
	}, w.apiSession)
	if err != nil {
		return err
	}
	log.Debug().Msgf("OpenZiti enrollment deleted: %s", enrollmentID)
	return nil
}

func (w *OpenZitiManagementWrapper) CreateEdgeRouter(routerName string) (string, error) {
	log.Debug().Msgf("Creating new OpenZiti edge router: %s", routerName)
	routerCreated, err := w.managementClient.API.EdgeRouter.CreateEdgeRouter(&edge_router.CreateEdgeRouterParams{
		EdgeRouter: &rest_model.EdgeRouterCreate{
			Name: &routerName,
		},
	}, w.apiSession)
	if err != nil {
		return "", err
	}
	edgeRouterID := routerCreated.Payload.Data.ID
	log.Debug().Msgf("OpenZiti edge router created with id: %s", edgeRouterID)

	router, err := w.managementClient.API.EdgeRouter.DetailEdgeRouter(&edge_router.DetailEdgeRouterParams{
		ID: edgeRouterID,
	}, w.apiSession)
	if err != nil {
		return "", err
	}
	return *router.Payload.Data.EnrollmentJWT, nil
}

func (w *OpenZitiManagementWrapper) CreateServicePolicy(svcPolicyName string, svcPolicySemantic rest_model.Semantic, svcPolicyType rest_model.DialBind, identityRoles, serviceRoles []string) error {
	log.Debug().Msgf("Creating new OpenZiti service policy: name=%s, type=%s, semantic=%s, identity-roles=%s, service-roles=%s", svcPolicyName, svcPolicyType, svcPolicySemantic, identityRoles, serviceRoles)
	_, err := w.managementClient.API.ServicePolicy.CreateServicePolicy(&service_policy.CreateServicePolicyParams{
		Policy: &rest_model.ServicePolicyCreate{
			Name:          &svcPolicyName,
			Semantic:      &svcPolicySemantic,
			IdentityRoles: identityRoles,
			ServiceRoles:  serviceRoles,
			Type:          &svcPolicyType,
		},
	}, w.apiSession)
	return err
}

func (w *OpenZitiManagementWrapper) CreateEdgeRouterPolicy(edgePolicyName string, edgePolicySemantic rest_model.Semantic, edgeRouterRoles, identityRoles []string) error {
	log.Debug().Msgf("Creating new OpenZiti edge router policy: name=%s, semantic=%s, edge-router-roles=%s, identity-roles=%s", edgePolicyName, edgePolicySemantic, edgeRouterRoles, identityRoles)
	_, err := w.managementClient.API.EdgeRouterPolicy.CreateEdgeRouterPolicy(&edge_router_policy.CreateEdgeRouterPolicyParams{
		Policy: &rest_model.EdgeRouterPolicyCreate{
			Name:            &edgePolicyName,
			Semantic:        &edgePolicySemantic,
			EdgeRouterRoles: edgeRouterRoles,
			IdentityRoles:   identityRoles,
		},
	}, w.apiSession)
	return err
}

func (w *OpenZitiManagementWrapper) CreateServiceEdgeRouterPolicy(policyName string, policySemantic rest_model.Semantic, edgeRouterRoles, serviceRoles []string) error {
	log.Debug().Msgf("Creating new OpenZiti service edge router policy: name=%s, semantic=%s, edge-router-roles=%s, service-roles=%s", policyName, policySemantic, edgeRouterRoles, serviceRoles)
	_, err := w.managementClient.API.ServiceEdgeRouterPolicy.CreateServiceEdgeRouterPolicy(&service_edge_router_policy.CreateServiceEdgeRouterPolicyParams{
		Policy: &rest_model.ServiceEdgeRouterPolicyCreate{
			Name:            &policyName,
			Semantic:        &policySemantic,
			EdgeRouterRoles: edgeRouterRoles,
			ServiceRoles:    serviceRoles,
		},
	}, w.apiSession)
	return err
}

func (w *OpenZitiManagementWrapper) CreateService(serviceName string, roleAttributes rest_model.Attributes) (string, error) {
	log.Debug().Msgf("Creating new OpenZiti service: name=%s, role-attributes=%s", serviceName, roleAttributes)
	encryptionRequired := true
	resp, err := w.managementClient.API.Service.CreateService(&service.CreateServiceParams{
		Service: &rest_model.ServiceCreate{
			Name:               &serviceName,
			EncryptionRequired: &encryptionRequired,
			RoleAttributes:     roleAttributes,
		},
	}, w.apiSession)
	if err != nil {
		return "", err
	}
	serviceID := resp.Payload.Data.ID
	log.Debug().Msgf("OpenZiti service created with id: %s", serviceID)
	return serviceID, err
}
