# lambdassmmiddlware
AWS Lambda middleware of ssm parameters middleware for Golang
don't need any extensions to use this middleware, get parameters using the SDK.

```go
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
		&lambdassmmiddleware.Config{
			Paths:          strings.Split(os.Getenv("SSMPATHS"), ","),
			Names:          strings.Split(os.Getenv("SSMNAMES"), ","),
			ContextKeyFunc: func(key string) interface{} { return ssmContextKey(key) },
			EnvPrefix:      "SSM_",
			CacheTTL:       24 * time.Hour,
			SetEnv:         true,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
	lambda.Start(handler)
}
```

see deteils [_example/](_example/)

```shell
$ make lambroll/invoke 
lambroll invoke
2023/03/28 17:15:38 [info] lambroll v0.14.2 with function.json
Enter JSON payloads for the invoking function into STDIN. (Type Ctrl-D to close.)
{}
{"bar":"bar values","env_bar":"bar values","env_foo":"foo values","env_hoge":"hoge values","foo":"foo values","hoge":"hoge values","tora":null}
2023/03/28 17:15:44 [info] StatusCode:200
2023/03/28 17:15:44 [info] ExecutionVersion:$LATEST
```

## LICENSE 

MIT
