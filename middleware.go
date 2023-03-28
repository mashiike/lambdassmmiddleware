package lambdassmmiddleware

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Songmu/flextime"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mashiike/lambdamiddleware"
)

type Config struct {
	Names          []string
	Client         *ssm.Client
	ContextKeyFunc func(key string) interface{}
	SetEnv         bool
	EnvPrefix      string
	withDecryption *bool
	CacheTTL       time.Duration

	mu             sync.RWMutex
	cache          map[string]string
	cacheFetchedAt map[string]time.Time
}

func (cfg *Config) WithDecryption(value bool) *Config {
	cfg.withDecryption = &value
	return cfg
}

func (cfg *Config) WithClient(client *ssm.Client) *Config {
	cfg.Client = client
	return cfg
}

func (cfg *Config) getFromChache(name string) (string, bool) {
	cfg.mu.RLock()
	defer cfg.mu.RUnlock()
	if cfg.cache == nil {
		return "", false
	}
	fetchedAt, ok := cfg.cacheFetchedAt[name]
	if !ok || fetchedAt.IsZero() {
		return "", false
	}
	lifetime := flextime.Since(fetchedAt)
	if lifetime < cfg.CacheTTL {
		return cfg.cache[name], true
	}
	return "", false
}

func (cfg *Config) setCache(name string, value string) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if cfg.cache == nil {
		cfg.cache = make(map[string]string)
		cfg.cacheFetchedAt = make(map[string]time.Time)
	}
	cfg.cacheFetchedAt[name] = flextime.Now()
	cfg.cache[name] = value
}

func Wrap(handler interface{}, cfg *Config) (lambda.Handler, error) {
	m, err := New(cfg)
	if err != nil {
		return nil, err
	}
	s := lambdamiddleware.NewStack(m)
	return s.Then(handler), nil
}

func New(cfg *Config) (lambdamiddleware.Middleware, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.Client == nil {
		awsCfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, fmt.Errorf("load default aws config:%w", err)
		}
		cfg.Client = ssm.NewFromConfig(awsCfg)
	}
	if cfg.withDecryption == nil {
		cfg = cfg.WithDecryption(true)
	}
	if cfg.ContextKeyFunc == nil {
		cfg.ContextKeyFunc = func(key string) interface{} {
			return key
		}
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 1 * time.Hour
	}
	return func(next lambda.Handler) lambda.Handler {
		return lambdamiddleware.HandlerFunc(func(ctx context.Context, payload []byte) ([]byte, error) {
			ctx, err := cfg.fetchParametersAndSecrets(ctx)
			if err != nil {
				return nil, err
			}
			return next.Invoke(ctx, payload)
		})
	}, nil
}

func (cfg *Config) fetchParametersAndSecrets(ctx context.Context) (context.Context, error) {
	setFunc := func(ctx context.Context, name string, value string) context.Context {
		if cfg.SetEnv {
			parts := strings.Split(name, "/")
			envKey := strings.ToUpper(cfg.EnvPrefix + parts[len(parts)-1])
			os.Setenv(envKey, value)
		}
		return context.WithValue(ctx, cfg.ContextKeyFunc(name), value)
	}
	refetchNames := make([]string, 0, len(cfg.Names))
	for _, name := range cfg.Names {
		if value, ok := cfg.getFromChache(name); ok {
			ctx = setFunc(ctx, name, value)
		} else {
			refetchNames = append(refetchNames, name)
		}
	}
	if len(refetchNames) == 0 {
		return ctx, nil
	}
	output, err := cfg.Client.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          refetchNames,
		WithDecryption: cfg.withDecryption,
	})
	if err != nil {
		return nil, err
	}
	for _, param := range output.Parameters {
		cfg.setCache(*param.Name, *param.Value)
		ctx = setFunc(ctx, *param.Name, *param.Value)
	}
	return ctx, nil
}
