package utils

import (
	"os"
	"strings"
)

var (
	Env                    = os.Getenv("ENV")
	Env_TracingServiceName = os.Getenv("TRACING_SERVICE_NAME")
	Env_OLTPEndpoint       = os.Getenv("OLTP_ENDPOINT")
	Env_ProxyEndpoint      = os.Getenv("PROXY_ENDPOINT")

	Env_TracingEnabled = os.Getenv("TRACING") == "1"

	CacheEnabled = os.Getenv("CACHE_ENABLED") == "1"
	// http://x:y,http://z:y,... MUST INCLUDE SELF! Only need to include self to cache as a single node
	CachePeers = strings.Split(os.Getenv("CACHE_PEERS"), ",")
	// http://x.x.x.x:yyyy
	CacheSelfAddr       = os.Getenv("CACHE_SELF_ADDR")
	Env_KeyCacheMB      = MustEnvOrDefaultInt64("KEY_CACHE_MB", 10_000_000)
	Env_KeyCacheSeconds = MustEnvOrDefaultInt64("KEY_CACHE_SECONDS", 300)

	Env_ControlPlaneAddr       = os.Getenv("CONTROL_PLANE_ADDR")
	Env_ControlPlaneAuthHeader = os.Getenv("CONTROL_PLANE_AUTH")

	Env_AWSService = os.Getenv("AWS_SERVICE")
	Env_Region     = os.Getenv("REGION")

	Env_ResourcePolicyCacheMB      = MustEnvOrDefaultInt64("RESOURCE_POLICIES_CACHE_MB", 10_000_000)
	Env_ResourcePolicyCacheSeconds = MustEnvOrDefaultInt64("RESOURCE_POLICIES_CACHE_SECONDS", 300)
)
