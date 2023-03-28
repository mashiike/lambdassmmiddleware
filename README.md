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
	"time"

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
				"foo":     ctx.Value(ssmContextKey("/lambdassmmiddleware/foo")),
				"bar":     ctx.Value(ssmContextKey("/lambdassmmiddleware/bar")),
				"tora":    ctx.Value(ssmContextKey("/lambdassmmiddleware/tora")),
			}, nil
		},
		&lambdassmmiddleware.Config{
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
2023/03/28 16:50:09 [info] lambroll v0.14.2 with function.json
Enter JSON payloads for the invoking function into STDIN. (Type Ctrl-D to close.)
{}
{"bar":"bar values","env_bar":"bar values","env_foo":"foo values","foo":"foo values","tora":null}
2023/03/28 16:50:20 [info] StatusCode:200
2023/03/28 16:50:20 [info] ExecutionVersion:$LATEST
```

## LICENSE 

MIT
