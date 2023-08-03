package lambda

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/aws/aws-xray-sdk-go/xray"
	_aws "github.com/brexhq/substation/internal/aws"
)

// New returns a configured Lambda client.

func New(cfg _aws.Config) *lambda.Lambda {
	conf, sess := _aws.New(cfg)

	c := lambda.New(sess, conf)
	if _, ok := os.LookupEnv("AWS_XRAY_DAEMON_ADDRESS"); ok {
		xray.AWS(c.Client)
	}

	return c
}

// API wraps the Lambda API interface.
type API struct {
	Client lambdaiface.LambdaAPI
}

// Setup creates a new Lambda client.
func (a *API) Setup(cfg _aws.Config) {
	a.Client = New(cfg)
}

// IsEnabled returns true if the client is enabled and ready for use.
func (a *API) IsEnabled() bool {
	return a.Client != nil
}

// Invoke is a convenience wrapper for synchronously invoking a Lambda function.
func (a *API) Invoke(ctx aws.Context, function string, payload []byte) (resp *lambda.InvokeOutput, err error) {
	resp, err = a.Client.InvokeWithContext(
		ctx,
		// TODO(v1.0.0): add ARN support
		&lambda.InvokeInput{
			FunctionName:   aws.String(function),
			InvocationType: aws.String("RequestResponse"),
			Payload:        payload,
		})

	if err != nil {
		return nil, fmt.Errorf("invoke function %s: %v", function, err)
	}

	return resp, nil
}
