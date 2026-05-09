package stack

import (
	"fmt"
	"os"
	"path/filepath"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	_jsii_ "github.com/aws/jsii-runtime-go"
)

func lambdaAssetPath(name string) *string {
	candidates := []string{
		filepath.Join("cdk", "dist", name),
		filepath.Join("dist", name),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return _jsii_.String(candidate)
		}
	}

	panic(fmt.Sprintf("lambda asset directory for %s not found; run `make build-lambdas`", name))
}

func uiAssetPath() *string {
	candidates := []string{
		filepath.Join("dist-ui"),
		filepath.Join("cdk", "dist-ui"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return _jsii_.String(candidate)
		}
	}

	panic("frontend asset directory not found; expected dist-ui")
}

func swaggerAssetPath() *string {
	candidates := []string{
		filepath.Join("src", "product-service", "docs"),
		filepath.Join("product-service", "docs"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return _jsii_.String(candidate)
		}
	}

	panic("swagger asset directory not found; expected src/product-service/docs")
}

// websiteBucketName returns the configured bucket name, or nil to let CDK
// generate a unique name. Using a hard-coded name causes "already exists" errors
// on repeated deployments when the previous bucket was not deleted.
func websiteBucketName() *string {
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return nil
	}

	return _jsii_.String(bucketName)
}

func deploymentEnv() *awscdk.Environment {
	account := os.Getenv("AWS_ACCOUNT_ID")
	if account == "" {
		account = os.Getenv("CDK_DEFAULT_ACCOUNT")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		region = os.Getenv("CDK_DEFAULT_REGION")
	}

	if account == "" || region == "" {
		return nil
	}

	return &awscdk.Environment{
		Account: _jsii_.String(account),
		Region:  _jsii_.String(region),
	}
}

// NewProductServiceStack creates the product service infrastructure in the
// existing EPAM shop website stack.
//
// Expected Lambda artifacts:
// - cdk/dist/get-products-list/bootstrap
// - cdk/dist/get-product-by-id/bootstrap
func NewProductServiceStack(scope constructs.Construct) awscdk.Stack {
	stack := awscdk.NewStack(scope, _jsii_.String("epam-shop-website-stack"), &awscdk.StackProps{
		StackName: _jsii_.String("epam-shop-website-stack"),
		Env:       deploymentEnv(),
	})

	websiteBucketProps := &awss3.BucketProps{
		AutoDeleteObjects: _jsii_.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
	}
	if name := websiteBucketName(); name != nil {
		websiteBucketProps.BucketName = name
	}
	websiteBucket := awss3.NewBucket(stack, _jsii_.String("epam-shop-bucket"), websiteBucketProps)

	distribution := awscloudfront.NewDistribution(stack, _jsii_.String("epam-shop-distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(websiteBucket, nil),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
		},
		DefaultRootObject: _jsii_.String("index.html"),
	})

	awss3deployment.NewBucketDeployment(stack, _jsii_.String("epam-shop-deployment-with-invalidation"), &awss3deployment.BucketDeploymentProps{
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(uiAssetPath(), nil),
		},
		DestinationBucket: websiteBucket,
		Distribution:      distribution,
		DistributionPaths: &[]*string{_jsii_.String("/*")},
	})

	awss3deployment.NewBucketDeployment(stack, _jsii_.String("swagger-deployment"), &awss3deployment.BucketDeploymentProps{
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(swaggerAssetPath(), nil),
		},
		DestinationBucket:    websiteBucket,
		DestinationKeyPrefix: _jsii_.String("docs/"),
		Distribution:         distribution,
		DistributionPaths:    &[]*string{_jsii_.String("/docs/*")},
	})

	productsTable := awsdynamodb.NewTable(stack, _jsii_.String("products-table"), &awsdynamodb.TableProps{
		TableName: _jsii_.String("products"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: _jsii_.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	stocksTable := awsdynamodb.NewTable(stack, _jsii_.String("stocks-table"), &awsdynamodb.TableProps{
		TableName: _jsii_.String("stocks"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: _jsii_.String("product_id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	listProductsFn := awslambda.NewFunction(stack, _jsii_.String("get-products-list-lambda"), &awslambda.FunctionProps{
		FunctionName: _jsii_.String("get-products-list"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      _jsii_.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(lambdaAssetPath("get-products-list"), nil),
		MemorySize:   _jsii_.Number(512),
		Timeout:      awscdk.Duration_Seconds(_jsii_.Number(10)),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE_NAME": productsTable.TableName(),
			"STOCKS_TABLE_NAME":   stocksTable.TableName(),
		},
	})

	getProductByIDFn := awslambda.NewFunction(stack, _jsii_.String("get-product-by-id-lambda"), &awslambda.FunctionProps{
		FunctionName: _jsii_.String("get-product-by-id"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      _jsii_.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(lambdaAssetPath("get-product-by-id"), nil),
		MemorySize:   _jsii_.Number(512),
		Timeout:      awscdk.Duration_Seconds(_jsii_.Number(10)),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE_NAME": productsTable.TableName(),
			"STOCKS_TABLE_NAME":   stocksTable.TableName(),
		},
	})

	productsTable.GrantReadData(listProductsFn)
	stocksTable.GrantReadData(listProductsFn)
	productsTable.GrantReadData(getProductByIDFn)
	stocksTable.GrantReadData(getProductByIDFn)

	createProductFn := awslambda.NewFunction(stack, _jsii_.String("create-product-lambda"), &awslambda.FunctionProps{
		FunctionName: _jsii_.String("create-product"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      _jsii_.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(lambdaAssetPath("create-product"), nil),
		MemorySize:   _jsii_.Number(512),
		Timeout:      awscdk.Duration_Seconds(_jsii_.Number(10)),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE_NAME": productsTable.TableName(),
			"STOCKS_TABLE_NAME":   stocksTable.TableName(),
		},
	})

	productsTable.GrantWriteData(createProductFn)
	stocksTable.GrantWriteData(createProductFn)

	api := awsapigateway.NewRestApi(stack, _jsii_.String("product-service-api"), &awsapigateway.RestApiProps{
		RestApiName: _jsii_.String("product-service"),
		Description: _jsii_.String("REST API for product service handlers"),
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowMethods: &[]*string{
				_jsii_.String("GET"),
				_jsii_.String("POST"),
				_jsii_.String("OPTIONS"),
			},
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
		},
		DeployOptions: &awsapigateway.StageOptions{
			StageName: _jsii_.String("prod"),
		},
	})

	productsResource := api.Root().AddResource(_jsii_.String("products"), nil)
	productsResource.AddMethod(
		_jsii_.String("GET"),
		awsapigateway.NewLambdaIntegration(listProductsFn, nil),
		nil,
	)
	productsResource.AddMethod(
		_jsii_.String("POST"),
		awsapigateway.NewLambdaIntegration(createProductFn, nil),
		nil,
	)

	productByIDResource := productsResource.AddResource(_jsii_.String("{id}"), nil)
	productByIDResource.AddMethod(
		_jsii_.String("GET"),
		awsapigateway.NewLambdaIntegration(getProductByIDFn, nil),
		nil,
	)

	awscdk.NewCfnOutput(stack, _jsii_.String("products-api-url"), &awscdk.CfnOutputProps{
		Value: api.Url(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("website-bucket-name"), &awscdk.CfnOutputProps{
		Value: websiteBucket.BucketName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("website-distribution-domain-name"), &awscdk.CfnOutputProps{
		Value: distribution.DistributionDomainName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-json-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/swagger.json"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-yaml-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/swagger.yaml"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("get-products-list-lambda-name"), &awscdk.CfnOutputProps{
		Value: listProductsFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("get-product-by-id-lambda-name"), &awscdk.CfnOutputProps{
		Value: getProductByIDFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("create-product-lambda-name"), &awscdk.CfnOutputProps{
		Value: createProductFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("products-table-name"), &awscdk.CfnOutputProps{
		Value: productsTable.TableName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("stocks-table-name"), &awscdk.CfnOutputProps{
		Value: stocksTable.TableName(),
	})

	return stack
}
