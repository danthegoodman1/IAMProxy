package http_server

import (
	"github.com/danthegoodman1/GoAPITemplate/tracing"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/rs/zerolog"
	"net/http"
)

func (s *HTTPServer) ProxyRequest(c *CustomContext) error {
	ctx, span := tracing.Tracer.Start(c.Request().Context(), "ProxyRequest")
	defer span.End()
	logger := zerolog.Ctx(ctx)

	logger.Debug().Str("authHeader", c.Request().Header.Get("Authorization")).Msg("proxying request")

	req, err := http.NewRequestWithContext(c.Request().Context(), c.Request().Method, utils.Env_ProxyEndpoint, nil)
	if err != nil {
		return c.InternalError(err, "error making new request for proxying")
	}

	// Copy headers
	headers := c.Request().Header.Clone()
	// headers.Del("Authorization")
	req.Header = headers

	req.Header.Add("x-req-id", c.RequestID)
	req.Header.Add("x-span-id", span.SpanContext().SpanID().String())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.InternalError(err, "error doing proxy request")
	}
	defer res.Body.Close()

	// Copy headers
	for name, headers := range res.Header {
		// Iterate all headers with one name (e.g. Content-Type)
		for _, hdr := range headers {
			c.Response().Header().Set(name, hdr)
		}
	}

	return c.Stream(res.StatusCode, res.Header.Get("content-type"), res.Body)
}
