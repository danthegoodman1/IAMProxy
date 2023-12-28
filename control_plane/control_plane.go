package control_plane

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/tracing"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/mailgun/groupcache/v2"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrHighStatusCode = errors.New("high status code")

	poolServer               *http.Server
	UserPoliciesGroupCache   *groupcache.Group
	ResourcePolicyGroupCache *groupcache.Group
)

func InitCache(ctx context.Context) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("initializing cache")
	pool := groupcache.NewHTTPPoolOpts(utils.CacheSelfAddr, &groupcache.HTTPPoolOptions{})

	// Add more peers to the cluster You MUST Ensure our instance is included in this list else
	// determining who owns the key across the cluster will not be consistent, and the pool won't
	// be able to determine if our instance owns the key.
	pool.Set(utils.CachePeers...)

	listenAddr := ":" + strings.Split(utils.CacheSelfAddr, ":")[2]
	poolServer = &http.Server{
		Addr:    listenAddr,
		Handler: pool,
	}

	// Start an HTTP server to listen for peer requests from the groupcache
	go func() {
		logger.Debug().Msgf("cache pool server listening on %s (HTTP)", listenAddr)
		if err := poolServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("error on pool server listen")
		}
	}()

	UserPoliciesGroupCache = groupcache.NewGroup("key", utils.Env_UserPoliciesCacheMB, groupcache.GetterFunc(
		func(ctx context.Context, fqdn string, dest groupcache.Sink) error {
			userPolicyBytes, err := getUserPolicies(ctx, fqdn)
			if err != nil {
				return fmt.Errorf("error in getUserPolicies: %w", err)
			}

			return dest.SetBytes(userPolicyBytes, time.Now().Add(time.Second*time.Duration(utils.UserPoliciesCacheSeconds)))
		},
	))

	ResourcePolicyGroupCache = groupcache.NewGroup("resource", utils.Env_ResourcePolicyCacheMB, groupcache.GetterFunc(
		func(ctx context.Context, fqdn string, dest groupcache.Sink) error {
			resourcePolicyBytes, err := getResourcePolicyBytes(ctx, fqdn)
			if err != nil {
				return fmt.Errorf("error in getResourcePolicyBytes: %w", err)
			}

			return dest.SetBytes(resourcePolicyBytes, time.Now().Add(time.Second*time.Duration(utils.Env_ResourcePolicyCacheSeconds)))
		},
	))
}

func StopCache(ctx context.Context) error {
	return poolServer.Shutdown(ctx)
}

// Bytes of routing.Config
func getUserPolicies(ctx context.Context, fqdn string) ([]byte, error) {
	ctx, span := tracing.Tracer.Start(ctx, "getUserPolicies")
	defer span.End()
	span.SetAttributes(attribute.String("fqdn", fqdn))
	span.SetAttributes(attribute.Bool("cached", false))

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/domains/%s/config", utils.Env_ControlPlaneAddr, fqdn), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", utils.Env_ControlPlaneAuthHeader))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}
	defer res.Body.Close()
	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("high status code %d - %s: %w", res.StatusCode, string(resBytes), ErrHighStatusCode)
	}

	return resBytes, nil
}

// Bytes of control_plane.GetCertRes
func getResourcePolicyBytes(ctx context.Context, fqdn string) ([]byte, error) {
	ctx, span := tracing.Tracer.Start(ctx, "getResourcePolicyBytes")
	defer span.End()
	span.SetAttributes(attribute.String("fqdn", fqdn))
	span.SetAttributes(attribute.Bool("cached", false))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/domains/%s/cert", utils.Env_ControlPlaneAddr, fqdn), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating new request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", utils.Env_ControlPlaneAuthHeader))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}
	defer res.Body.Close()
	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	if res.StatusCode > 299 {
		return nil, fmt.Errorf("high status code %d - %s: %w", res.StatusCode, string(resBytes), ErrHighStatusCode)
	}

	return resBytes, nil
}

func GetUserPolicies(ctx context.Context, fqdn string) (*routing.Config, error) {
	ctx, span := tracing.Tracer.Start(ctx, "GetUserPolicies")
	defer span.End()
	span.SetAttributes(attribute.String("fqdn", fqdn))

	var b []byte
	var err error
	if utils.CacheEnabled {
		err = FQDNGroupCache.Get(ctx, fqdn, groupcache.AllocatingByteSliceSink(&b))
		if err != nil {
			return nil, fmt.Errorf("error getting from groupcache: %w", err)
		}
	} else {
		b, err = getUserPolicies(ctx, fqdn)
		if err != nil {
			return nil, fmt.Errorf("error in getUserPolicies: %w", err)
		}
	}

	fmt.Printf("got fqdn bytes %s\n", b)

	var config routing.Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal: %w", err)
	}

	return &config, nil
}

func GetResourcePolicy(ctx context.Context, fqdn string) (*tls.Certificate, error) {
	ctx, span := tracing.GildraTracer.Start(ctx, "GetResourcePolicy")
	defer span.End()
	span.SetAttributes(attribute.String("fqdn", fqdn))

	var b []byte
	var err error
	if utils.CacheEnabled {
		err = CertGroupCache.Get(ctx, fqdn, groupcache.AllocatingByteSliceSink(&b))
		if err != nil {
			return nil, fmt.Errorf("error getting from groupcache: %w", err)
		}
	} else {
		b, err = getResourcePolicyBytes(ctx, fqdn)
		if err != nil {
			return nil, fmt.Errorf("error in getResourcePolicyBytes: %w", err)
		}
	}

	var cert GetCertRes
	err = json.Unmarshal(b, &cert)
	if err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal: %w", err)
	}

	c, err := cert.GetCert()
	if err != nil {
		return nil, fmt.Errorf("error in cert.GetCert(): %w", err)
	}

	return c, nil
}
