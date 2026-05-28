package stack

import (
	"os"
	"strings"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	_jsii_ "github.com/aws/jsii-runtime-go"
)

func basicAuthorizerPassword() string {
	password := strings.TrimSpace(os.Getenv("siarhei_kurachou"))
	if password == "" {
		return "TEST_PASSWORD"
	}

	return password
}

// NewAuthorizationServiceStack creates resources for authorization-service.
func NewAuthorizationServiceStack(scope constructs.Construct) (awscdk.Stack, *string) {
	stack := awscdk.NewStack(scope, _jsii_.String("authorization-service-stack"), &awscdk.StackProps{
		StackName: _jsii_.String("authorization-service-stack"),
		Env:       deploymentEnv(),
	})

	basicAuthorizerFn := awslambda.NewFunction(stack, _jsii_.String("basic-authorizer-lambda"), &awslambda.FunctionProps{
		FunctionName: _jsii_.String("basicAuthorizer"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      _jsii_.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(lambdaAssetPath("basic-authorizer"), nil),
		MemorySize:   _jsii_.Number(512),
		Timeout:      awscdk.Duration_Seconds(_jsii_.Number(10)),
		Environment: &map[string]*string{
			"siarhei_kurachou": _jsii_.String(basicAuthorizerPassword()),
		},
	})

	awslambda.NewCfnPermission(stack, _jsii_.String("basic-authorizer-apigateway-permission"), &awslambda.CfnPermissionProps{
		Action:       _jsii_.String("lambda:InvokeFunction"),
		FunctionName: basicAuthorizerFn.FunctionName(),
		Principal:    _jsii_.String("apigateway.amazonaws.com"),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("basic-authorizer-lambda-arn"), &awscdk.CfnOutputProps{
		Value: basicAuthorizerFn.FunctionArn(),
	})

	return stack, basicAuthorizerFn.FunctionArn()
}
