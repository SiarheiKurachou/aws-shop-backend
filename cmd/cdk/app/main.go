package main

import (
	"aws-shop-backend/cmd/cdk/stack"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	_jsii_ "github.com/aws/jsii-runtime-go"
)

func main() {
	defer _jsii_.Close()

	app := awscdk.NewApp(nil)
	_, basicAuthorizerArn := stack.NewAuthorizationServiceStack(app)
	stack.NewProductServiceStack(app, basicAuthorizerArn)
	stack.NewBFFServiceStack(app)
	app.Synth(nil)
}
