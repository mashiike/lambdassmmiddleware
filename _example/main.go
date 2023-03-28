package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mashiike/lambdassmmiddleware"
)

type ssmContextKey string

func main() {
	handler, err := lambdassmmiddleware.Wrap(
		&lambdassmmiddleware.Config{
			Paths:          strings.Split(os.Getenv("SSMPATHS"), ","),
			Names:          strings.Split(os.Getenv("SSMNAMES"), ","),
			ContextKeyFunc: func(key string) interface{} { return ssmContextKey(key) },
			EnvPrefix:      "SSM_",
			CacheTTL:       24 * time.Hour,
			SetEnv:         true,
		},
		func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
			return map[string]interface{}{
				"env_hoge": os.Getenv("SSM_HOGE"),
				"env_foo":  os.Getenv("SSM_FOO"),
				"env_bar":  os.Getenv("SSM_BAR"),
				"hoge":     ctx.Value(ssmContextKey("/lambdassmmiddleware/paths/hoge")),
				"foo":      ctx.Value(ssmContextKey("/lambdassmmiddleware/foo")),
				"bar":      ctx.Value(ssmContextKey("/lambdassmmiddleware/bar")),
				"tora":     ctx.Value(ssmContextKey("/lambdassmmiddleware/tora")),
			}, nil
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
	lambda.Start(handler)
}
