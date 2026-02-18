package config

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	env "github.com/caarlos0/env/v11"
)

// This project was ported from Node.js, hence the NodeEnv for compatibility
type Config struct {
	NodeEnv              string
	Port                 int
	RedisURL             string
	Channels             string
	HeartbeatIntervalSec int
	SendRawRedisMessages bool
}

func LoadFromEnv() (Config, error) {
	type rawConfig struct {
		NodeEnv              string     `env:"NODE_ENV" envDefault:"development"`
		Port                 jsInt      `env:"PORT" envDefault:"3000"`
		RedisURL             validURL   `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
		Channels             string     `env:"CHANNELS" envDefault:"*"`
		HeartbeatIntervalSec jsInt      `env:"HEARTBEAT_INTERVAL" envDefault:"30"`
		SendRawRedisMessages strictBool `env:"SEND_RAW_REDIS_MESSAGES" envDefault:"true"`
	}

	parsed := rawConfig{}
	if err := env.ParseWithOptions(&parsed, env.Options{
		Environment: nonEmptyEnvironment(os.Environ()),
	}); err != nil {
		return Config{}, err
	}

	return Config{
		NodeEnv:              parsed.NodeEnv,
		Port:                 int(parsed.Port),
		RedisURL:             string(parsed.RedisURL),
		Channels:             parsed.Channels,
		HeartbeatIntervalSec: int(parsed.HeartbeatIntervalSec),
		SendRawRedisMessages: bool(parsed.SendRawRedisMessages),
	}, nil
}

type jsInt int

func (v *jsInt) UnmarshalText(text []byte) error {
	parsed, err := parseIntBase10LikeJS(string(text))
	if err != nil {
		return err
	}

	*v = jsInt(parsed)
	return nil
}

type strictBool bool

func (v *strictBool) UnmarshalText(text []byte) error {
	*v = strictBool(string(text) == "true")
	return nil
}

type validURL string

func (v *validURL) UnmarshalText(text []byte) error {
	raw := string(text)
	if err := validateURL(raw); err != nil {
		return err
	}

	*v = validURL(raw)
	return nil
}

func nonEmptyEnvironment(entries []string) map[string]string {
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if value == "" {
			continue
		}
		result[key] = value
	}

	return result
}

var leadingBase10IntPattern = regexp.MustCompile(`^[\t\n\r\f\v ]*([+-]?\d+)`)

func parseIntBase10LikeJS(input string) (int, error) {
	matches := leadingBase10IntPattern.FindStringSubmatch(input)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid integer: %q", input)
	}

	n, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %q", input)
	}

	return n, nil
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid url: %q", raw)
	}

	return nil
}
