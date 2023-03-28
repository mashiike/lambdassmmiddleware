package lambdassmmiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Songmu/flextime"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	restore := flextime.Set(time.Date(2023, 04, 01, 01, 00, 00, 00, time.UTC))
	defer restore()

	cases := []struct {
		casename  string
		handler   interface{}
		cfg       *Config
		payload   []byte
		apiOutput *ssm.GetParametersOutput
		apiErr    error
		output    string
		errString string
	}{
		{
			casename: "success",
			cfg: &Config{
				Names: []string{"hoge", "fuga"},
				ContextKeyFunc: func(key string) interface{} {
					return "ssm:" + key
				},
			},
			handler: func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
				require.EqualValues(t, `"hoge"`, string(payload))
				return map[string]interface{}{
					"hoge": ctx.Value("ssm:hoge"),
					"fuga": ctx.Value("ssm:fuga"),
				}, nil
			},
			apiOutput: &ssm.GetParametersOutput{
				Parameters: []types.Parameter{
					{
						Name:  aws.String("hoge"),
						Value: aws.String("hoge_dummy_value"),
					},
					{
						Name:  aws.String("fuga"),
						Value: aws.String("fuga_dummy_value"),
					},
				},
			},
			payload: []byte(`"hoge"`),
			output:  `{"hoge":"hoge_dummy_value","fuga":"fuga_dummy_value"}`,
		},
		{
			casename: "success with set env",
			cfg: &Config{
				Names:     []string{"hoge", "fuga"},
				SetEnv:    true,
				EnvPrefix: "SSM_",
			},
			handler: func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
				require.EqualValues(t, `"hoge"`, string(payload))
				return map[string]interface{}{
					"hoge": os.Getenv("SSM_HOGE"),
					"fuga": os.Getenv("SSM_FUGA"),
				}, nil
			},
			apiOutput: &ssm.GetParametersOutput{
				Parameters: []types.Parameter{
					{
						Name:  aws.String("hoge"),
						Value: aws.String("hoge_dummy_value"),
					},
					{
						Name:  aws.String("fuga"),
						Value: aws.String("fuga_dummy_value"),
					},
				},
			},
			payload: []byte(`"hoge"`),
			output:  `{"hoge":"hoge_dummy_value","fuga":"fuga_dummy_value"}`,
		},
		{
			casename: "success get from cache",
			cfg: &Config{
				Names: []string{"hoge", "fuga"},
				ContextKeyFunc: func(key string) interface{} {
					return "ssm:" + key
				},
				CacheTTL: time.Hour,
				cache: map[string]string{
					"hoge": "hoge_dummy_value",
					"fuga": "fuga_dummy_value",
				},
				cacheFetchedAt: map[string]time.Time{
					"hoge": time.Date(2023, 04, 01, 00, 59, 00, 00, time.UTC),
					"fuga": time.Date(2023, 04, 01, 00, 59, 00, 00, time.UTC),
				},
			},
			handler: func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
				require.EqualValues(t, `"hoge"`, string(payload))
				return map[string]interface{}{
					"hoge": ctx.Value("ssm:hoge"),
					"fuga": ctx.Value("ssm:fuga"),
				}, nil
			},
			apiOutput: &ssm.GetParametersOutput{
				Parameters: []types.Parameter{},
			},
			payload: []byte(`"hoge"`),
			output:  `{"hoge":"hoge_dummy_value","fuga":"fuga_dummy_value"}`,
		},
		{
			casename: "success refetch",
			cfg: &Config{
				Names: []string{"hoge", "fuga"},
				ContextKeyFunc: func(key string) interface{} {
					return "ssm:" + key
				},
				CacheTTL: time.Hour,

				cache: map[string]string{
					"hoge": "hoge_dummy_value",
					"fuga": "fuga_dummy_value",
				},
				cacheFetchedAt: map[string]time.Time{
					"hoge": time.Date(2023, 03, 30, 00, 00, 00, 00, time.UTC),
					"fuga": time.Date(2023, 03, 30, 00, 00, 00, 00, time.UTC),
				},
			},
			handler: func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
				require.EqualValues(t, `"hoge"`, string(payload))
				return map[string]interface{}{
					"hoge": ctx.Value("ssm:hoge"),
					"fuga": ctx.Value("ssm:fuga"),
				}, nil
			},
			apiOutput: &ssm.GetParametersOutput{
				Parameters: []types.Parameter{
					{
						Name:  aws.String("hoge"),
						Value: aws.String("hoge_dummy_value_v2"),
					},
					{
						Name:  aws.String("fuga"),
						Value: aws.String("fuga_dummy_value_v2"),
					},
				},
			},
			payload: []byte(`"hoge"`),
			output:  `{"hoge":"hoge_dummy_value_v2","fuga":"fuga_dummy_value_v2"}`,
		},
		{
			casename: "api error",
			cfg: &Config{
				Names: []string{"hoge", "fuga"},
				ContextKeyFunc: func(key string) interface{} {
					return "ssm:" + key
				},
			},
			handler: func(ctx context.Context, payload json.RawMessage) (interface{}, error) {
				return nil, errors.New("unexpected exection")
			},
			apiErr:    errors.New("unknown error"),
			payload:   []byte(`"hoge"`),
			errString: "operation error SSM: GetParameters, unknown error",
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case.%d:%s", i+1, c.casename), func(t *testing.T) {
			client := ssm.New(
				ssm.Options{
					Region: "api-northeast-1",
				},
				ssm.WithAPIOptions(func(stack *middleware.Stack) error {
					return stack.Finalize.Add(
						middleware.FinalizeMiddlewareFunc(
							"test",
							func(context.Context, middleware.FinalizeInput, middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
								return middleware.FinalizeOutput{
									Result: c.apiOutput,
								}, middleware.Metadata{}, c.apiErr
							},
						),
						middleware.Before,
					)
				}),
			)
			h, err := Wrap(c.handler, c.cfg.WithClient(client))
			require.NoError(t, err)
			output, err := h.Invoke(context.Background(), c.payload)
			if c.errString == "" {
				require.NoError(t, err)
				require.JSONEq(t, c.output, string(output))
			} else {
				require.EqualError(t, err, c.errString)
			}
		})
	}
}
