package main

import (
	"aws-shop-backend/cdk/stack"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	_jsii_ "github.com/aws/jsii-runtime-go"
)

func main() {
	defer _jsii_.Close()

	app := awscdk.NewApp(nil)
	stack.NewProductServiceStack(app)
	app.Synth(nil)
}
