# lambdassmmiddlware
AWS Lambda middleware of ssm parameters middleware for Golang


```go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mashiike/lambdassmmiddleware"
)

type ssmContextKey string

func main() {
	handler, err := lambdassmmiddleware.Wrap(
		func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
			return map[string]interface{}{
				"env_foo": os.Getenv("SSM_FOO"),
				"env_bar": os.Getenv("SSM_BAR"),
				"foo":     ctx.Value(ssmContextKey("/lambdassmmiddleware-examples/foo")),
				"bar":     ctx.Value(ssmContextKey("/lambdassmmiddleware-examples/bar")),
				"tora":    ctx.Value(ssmContextKey("/lambdassmmiddleware-examples/tora")),
			}, nil
		},
		lambdassmmiddleware.Config{
			Names:          strings.Split(os.Getenv("SSMNAMES"), ","),
			ContextKeyFunc: func(key string) interface{} { return ssmContextKey(key) },
			EnvPrefix:      "SSM_",
			SetEnv:         true,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
	lambda.Start(handler)
}
```

see deteils [exampels/parameters-and-secrets](_examples/parameters-and-secrets)

## LICENSE 

MIT
