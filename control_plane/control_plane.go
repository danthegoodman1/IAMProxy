package control_plane

import (
	"context"
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
	KeyGroupCache            *groupcache.Group
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

	KeyGroupCache = groupcache.NewGroup("key", utils.Env_KeyCacheMB, groupcache.GetterFunc(
		func(ctx context.Context, fqdn string, dest groupcache.Sink) error {
			userBytes, err := getKey(ctx, fqdn)
			if err != nil {
				return fmt.Errorf("error in getKey: %w", err)
			}

			return dest.SetBytes(userBytes, time.Now().Add(time.Second*time.Duration(utils.Env_KeyCacheSeconds)))
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
func getKey(ctx context.Context, keyID string) ([]byte, error) {
	ctx, span := tracing.Tracer.Start(ctx, "getKey")
	defer span.End()
	span.SetAttributes(attribute.String("keyID", keyID))
	span.SetAttributes(attribute.Bool("cached", false))

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/key/%s", utils.Env_ControlPlaneAddr, keyID), nil)
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

func GetKey(ctx context.Context, keyID string) (*Key, error) {
	ctx, span := tracing.Tracer.Start(ctx, "GetKey")
	defer span.End()
	span.SetAttributes(attribute.String("keyID", keyID))

	var b []byte
	var err error
	if utils.CacheEnabled {
		err = KeyGroupCache.Get(ctx, keyID, groupcache.AllocatingByteSliceSink(&b))
		if err != nil {
			return nil, fmt.Errorf("error getting from groupcache: %w", err)
		}
	} else {
		b, err = getKey(ctx, keyID)
		if err != nil {
			return nil, fmt.Errorf("error in getKey: %w", err)
		}
	}

	var user Key
	err = json.Unmarshal(b, &user)
	if err != nil {
		return nil, fmt.Errorf("error in json.Unmarshal: %w", err)
	}

	return &user, nil
}
