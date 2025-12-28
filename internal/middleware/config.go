package middleware

import (
	"context"
	"net/http"
	"tiny-bitly/internal/apperrors"
	"tiny-bitly/internal/config"
)

const configKey contextType = "config"

// Injects the provided app config into the request context.
func ConfigMiddleware(next http.Handler, config config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), configKey, config)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// Extracts the config from the context, or ErrConfigurationMissing if the
// config is not present in the context.
func GetConfigFromContext(ctx context.Context) (*config.Config, error) {
	if config, ok := ctx.Value(configKey).(config.Config); ok {
		return &config, nil
	}
	return nil, apperrors.ErrConfigurationMissing
}

// Sets the provided config in the provided context for use in tests.
func SetConfigInContextForTesting(ctx context.Context, config config.Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}
