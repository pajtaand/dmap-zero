package wrapper

import (
	"context"
	"fmt"
	"net"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/edge"
	"github.com/openziti/sdk-golang/ziti/enroll"
	"github.com/rs/zerolog/log"
)

const (
	serviceTerminatorLimit = 99999
)

type OpenZitiClientWrapperConfig struct {
	KeyAlg ziti.KeyAlgVar
}

type OpenZitiClientWrapper struct {
	cfg     *OpenZitiClientWrapperConfig
	zitiCfg *ziti.Config
	ztx     ziti.Context
}

func NewOpenZitiClientWrapperFromToken(cfg *OpenZitiClientWrapperConfig, enrollmentToken string) (*OpenZitiClientWrapper, error) {
	log.Debug().Msg("Creating new OpenZitiClientWrapper from token")

	w := &OpenZitiClientWrapper{
		cfg: cfg,
	}

	zitiCfg, err := w.enroll(enrollmentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll: %v", err)
	}
	w.zitiCfg = zitiCfg

	ztx, err := w.authenticate(zitiCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %v", err)
	}
	w.ztx = ztx

	return w, nil
}

func NewOpenZitiClientWrapperFromConfig(cfg *OpenZitiClientWrapperConfig, zitiCfg *ziti.Config) (*OpenZitiClientWrapper, error) {
	log.Debug().Msg("Creating new OpenZitiClientWrapper from config")

	w := &OpenZitiClientWrapper{
		cfg:     cfg,
		zitiCfg: zitiCfg,
	}

	ztx, err := w.authenticate(zitiCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %v", err)
	}
	w.ztx = ztx

	return w, nil
}

func (w *OpenZitiClientWrapper) Listen(service string) (edge.Listener, error) {
	log.Debug().Msgf("Listening for OpenZiti service: %s", service)
	listener, err := w.ztx.Listen(service)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %v", err)
	}
	return listener, nil
}

func (w *OpenZitiClientWrapper) ListenWithOptions(service string, options *ziti.ListenOptions) (edge.Listener, error) {
	log.Debug().Msgf("Listening for OpenZiti service with options: %s, %v", service, options)
	listener, err := w.ztx.ListenWithOptions(service, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %v", err)
	}
	return listener, nil
}

func (w *OpenZitiClientWrapper) GetContextDialer() func(context.Context, string) (net.Conn, error) {
	log.Debug().Msg("Getting OpenZiti context dialer")
	return func(ctx context.Context, s string) (net.Conn, error) {
		return w.ztx.Dial(s)
	}
}

func (w *OpenZitiClientWrapper) GetContextDialerWithOptions(options *ziti.DialOptions) func(context.Context, string) (net.Conn, error) {
	log.Debug().Msgf("Getting OpenZiti context dialer with options: %v", options)
	return func(ctx context.Context, s string) (net.Conn, error) {
		return w.ztx.DialWithOptions(s, options)
	}
}

func (w *OpenZitiClientWrapper) GetServiceTerminators(service string) ([]string, error) {
	log.Debug().Msgf("Getting OpenZiti service terminators for: %s", service)
	identities := []string{}
	terminators, _, err := w.ztx.GetServiceTerminators(service, 0, serviceTerminatorLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to get service terminators: %v", err)
	}
	for _, terminator := range terminators {
		identities = append(identities, *terminator.Identity)
	}
	log.Debug().Msgf("Found %d OpenZiti service terminators for: %s: %+v", len(identities), service, identities)
	return identities, nil
}

func (w *OpenZitiClientWrapper) GetIdentity() (string, error) {
	log.Debug().Msg("Getting OpenZiti identity")
	identity, err := w.ztx.GetCurrentIdentity()
	if err != nil {
		return "", fmt.Errorf("failed to get agent's identity: %v", err)
	}
	identityName := *identity.Name
	log.Debug().Msgf("OpenZiti identity: %s", identityName)
	return identityName, nil
}

func (w *OpenZitiClientWrapper) GetOpenZitiConfig() *ziti.Config {
	return w.zitiCfg
}

func (w *OpenZitiClientWrapper) enroll(enrollmentToken string) (*ziti.Config, error) {
	log.Debug().Msg("Enrolling to OpenZiti using JWT")

	tkn, _, err := enroll.ParseToken(enrollmentToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	cfg, err := enroll.Enroll(enroll.EnrollmentFlags{
		KeyAlg: w.cfg.KeyAlg,
		Token:  tkn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity: %v", err)
	}

	return cfg, nil
}

func (w *OpenZitiClientWrapper) authenticate(cfg *ziti.Config) (ziti.Context, error) {
	log.Debug().Msg("Authenticating to OpenZiti with configuration")

	ztx, err := ziti.NewContext(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new ziti context: %v", err)
	}

	if err := ztx.Authenticate(); err != nil {
		return nil, fmt.Errorf("failed to authenticate: %v", err)
	}
	return ztx, nil
}
