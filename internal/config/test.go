package config

// Returns a config object with defaults overridden by any defined fields in the
// provided config. Intended for use in tests only.
func GetTestConfig(cfg Config) Config {
	newCfg := GetDefaultConfig()

	if cfg.APIPort != 0 {
		newCfg.APIPort = cfg.APIPort
	}
	if cfg.APIHostname != "" {
		newCfg.APIHostname = cfg.APIHostname
	}
	if cfg.RateLimitRequestsPerSecond != 0 {
		newCfg.RateLimitRequestsPerSecond = cfg.RateLimitRequestsPerSecond
	}
	if cfg.RateLimitBurst != 0 {
		newCfg.RateLimitBurst = cfg.RateLimitBurst
	}
	if cfg.MaxAliasLength != 0 {
		newCfg.MaxAliasLength = cfg.MaxAliasLength
	}
	if cfg.MaxTriesCreateShortCode != 0 {
		newCfg.MaxTriesCreateShortCode = cfg.MaxTriesCreateShortCode
	}
	if cfg.MaxURLLength != 0 {
		newCfg.MaxURLLength = cfg.MaxURLLength
	}
	if cfg.ShortCodeLength != 0 {
		newCfg.ShortCodeLength = cfg.ShortCodeLength
	}
	if cfg.IdleTimeout != 0 {
		newCfg.IdleTimeout = cfg.IdleTimeout
	}
	if cfg.ReadTimeout != 0 {
		newCfg.ReadTimeout = cfg.ReadTimeout
	}
	if cfg.RequestTimeout != 0 {
		newCfg.RequestTimeout = cfg.RequestTimeout
	}
	if cfg.ShortCodeTTL != 0 {
		newCfg.ShortCodeTTL = cfg.ShortCodeTTL
	}
	if cfg.ShutdownTimeout != 0 {
		newCfg.ShutdownTimeout = cfg.ShutdownTimeout
	}
	if cfg.WriteTimeout != 0 {
		newCfg.WriteTimeout = cfg.WriteTimeout
	}

	return newCfg
}
