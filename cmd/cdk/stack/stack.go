package stack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3notifications"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
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

func swaggerAssetPath(service string) *string {
	candidates := []string{
		filepath.Join("dist", "docs", service),
		filepath.Join("cdk", "dist", "docs", service),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return _jsii_.String(candidate)
		}
	}

	panic(fmt.Sprintf(
		"swagger asset directory for %s not found; expected dist/docs/%s (run `make build-swagger`)",
		service,
		service,
	))
}

func swaggerDeploymentSources(path *string, service string) *[]awss3deployment.ISource {
	// Add a tiny synthetic source that changes each synth so the deployment
	// runs on every `cdk deploy`, restoring docs even if they were deleted manually.
	return &[]awss3deployment.ISource{
		awss3deployment.Source_Asset(path, nil),
		awss3deployment.Source_Data(
			_jsii_.String(".deployment-trigger"),
			_jsii_.String(fmt.Sprintf("service=%s deployed_at=%s", service, time.Now().UTC().Format(time.RFC3339Nano))),
			nil,
		),
	}
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

func existingCognitoUserPoolID() string {
	return strings.TrimSpace(os.Getenv("EXISTING_COGNITO_USER_POOL_ID"))
}

func existingCognitoUserPoolClientID() string {
	return strings.TrimSpace(os.Getenv("EXISTING_COGNITO_USER_POOL_CLIENT_ID"))
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

// NewProductServiceStack creates the product and import service infrastructure in the
// existing EPAM shop website stack.
//
// Expected Lambda artifacts:
// - cdk/dist/get-products-list/bootstrap
// - cdk/dist/get-product-by-id/bootstrap
func NewProductServiceStack(scope constructs.Construct, basicAuthorizerArn *string) awscdk.Stack {
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

	spaRewriteFunction := awscloudfront.NewFunction(stack, _jsii_.String("spa-rewrite-function"), &awscloudfront.FunctionProps{
		Code: awscloudfront.FunctionCode_FromInline(_jsii_.String(`function handler(event) {
	var request = event.request;
	var uri = request.uri;

	if (uri === '/docs' || uri.indexOf('/docs/') === 0) {
		return request;
	}

	if (uri.indexOf('.') === -1) {
		request.uri = '/index.html';
	}

	return request;
}`)),
	})

	distribution := awscloudfront.NewDistribution(stack, _jsii_.String("epam-shop-distribution"), &awscloudfront.DistributionProps{
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin:               awscloudfrontorigins.S3BucketOrigin_WithOriginAccessControl(websiteBucket, nil),
			ViewerProtocolPolicy: awscloudfront.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			FunctionAssociations: &[]*awscloudfront.FunctionAssociation{
				{
					EventType: awscloudfront.FunctionEventType_VIEWER_REQUEST,
					Function:  spaRewriteFunction,
				},
			},
		},
		DefaultRootObject: _jsii_.String("index.html"),
	})

	allowedOrigin := awscdk.Fn_Join(_jsii_.String(""), &[]*string{
		_jsii_.String("https://"),
		distribution.DistributionDomainName(),
	})

	mainUIDeployment := awss3deployment.NewBucketDeployment(stack, _jsii_.String("epam-shop-deployment-with-invalidation"), &awss3deployment.BucketDeploymentProps{
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(uiAssetPath(), nil),
		},
		DestinationBucket: websiteBucket,
		Prune:             _jsii_.Bool(false),
		Exclude: &[]*string{
			_jsii_.String("docs/*"),
			_jsii_.String("docs/**/*"),
		},
		Distribution:      distribution,
		DistributionPaths: &[]*string{_jsii_.String("/*")},
	})

	productSwaggerDocsPath := swaggerAssetPath("product-service")
	importSwaggerDocsPath := swaggerAssetPath("import-service")
	authorizationSwaggerDocsPath := swaggerAssetPath("authorization-service")

	// Namespaced product docs endpoint:
	// https://<distribution>/docs/product-service/swagger.json
	productSwaggerDeployment := awss3deployment.NewBucketDeployment(stack, _jsii_.String("swagger-product-service-deployment"), &awss3deployment.BucketDeploymentProps{
		Sources:              swaggerDeploymentSources(productSwaggerDocsPath, "product-service"),
		DestinationBucket:    websiteBucket,
		DestinationKeyPrefix: _jsii_.String("docs/product-service/"),
		Distribution:         distribution,
		DistributionPaths:    &[]*string{_jsii_.String("/docs/product-service/*")},
	})

	// Namespaced import docs endpoint:
	// https://<distribution>/docs/import-service/swagger.json
	importSwaggerDeployment := awss3deployment.NewBucketDeployment(stack, _jsii_.String("swagger-import-service-deployment"), &awss3deployment.BucketDeploymentProps{
		Sources:              swaggerDeploymentSources(importSwaggerDocsPath, "import-service"),
		DestinationBucket:    websiteBucket,
		DestinationKeyPrefix: _jsii_.String("docs/import-service/"),
		Distribution:         distribution,
		DistributionPaths:    &[]*string{_jsii_.String("/docs/import-service/*")},
	})

	// Namespaced authorization docs endpoint:
	// https://<distribution>/docs/authorization-service/swagger.json
	authorizationSwaggerDeployment := awss3deployment.NewBucketDeployment(stack, _jsii_.String("swagger-authorization-service-deployment"), &awss3deployment.BucketDeploymentProps{
		Sources:              swaggerDeploymentSources(authorizationSwaggerDocsPath, "authorization-service"),
		DestinationBucket:    websiteBucket,
		DestinationKeyPrefix: _jsii_.String("docs/authorization-service/"),
		Distribution:         distribution,
		DistributionPaths:    &[]*string{_jsii_.String("/docs/authorization-service/*")},
	})

	productSwaggerDeployment.Node().AddDependency(mainUIDeployment)
	importSwaggerDeployment.Node().AddDependency(mainUIDeployment)
	authorizationSwaggerDeployment.Node().AddDependency(mainUIDeployment)

	importsBucket := awss3.NewBucket(stack, _jsii_.String("imports-bucket"), &awss3.BucketProps{
		AutoDeleteObjects: _jsii_.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		Cors: &[]*awss3.CorsRule{
			{
				AllowedMethods: &[]awss3.HttpMethods{awss3.HttpMethods_PUT},
				AllowedOrigins: &[]*string{_jsii_.String("*")},
				AllowedHeaders: &[]*string{_jsii_.String("*")},
			},
		},
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
			"ALLOWED_ORIGIN":      allowedOrigin,
		},
	})

	importProductsFileFn := awslambda.NewFunction(stack, _jsii_.String("import-products-file-lambda"), &awslambda.FunctionProps{
		FunctionName: _jsii_.String("import-products-file"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      _jsii_.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(lambdaAssetPath("import-products-file"), nil),
		MemorySize:   _jsii_.Number(512),
		Timeout:      awscdk.Duration_Seconds(_jsii_.Number(10)),
		Environment: &map[string]*string{
			"IMPORT_BUCKET_NAME": importsBucket.BucketName(),
		},
	})

	importFileParserFn := awslambda.NewFunction(stack, _jsii_.String("import-file-parser-lambda"), &awslambda.FunctionProps{
		FunctionName: _jsii_.String("import-file-parser"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      _jsii_.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(lambdaAssetPath("import-file-parser"), nil),
		MemorySize:   _jsii_.Number(512),
		Timeout:      awscdk.Duration_Seconds(_jsii_.Number(10)),
		Environment: &map[string]*string{
			"IMPORT_BUCKET_NAME": importsBucket.BucketName(),
		},
	})

	catalogItemsQueue := awssqs.NewQueue(stack, _jsii_.String("catalog-items-queue"), &awssqs.QueueProps{
		QueueName: _jsii_.String("catalogItemsQueue"),
	})

	createProductTopic := awssns.NewTopic(stack, _jsii_.String("create-product-topic"), &awssns.TopicProps{
		TopicName: _jsii_.String("createProductTopic"),
	})

	budgetNotificationEmail := strings.TrimSpace(os.Getenv("PRODUCT_NOTIFICATION_EMAIL"))
	if budgetNotificationEmail == "" {
		budgetNotificationEmail = "your-email@example.com"
	}

	premiumNotificationEmail := strings.TrimSpace(os.Getenv("PRODUCT_NOTIFICATION_EMAIL_PREMIUM"))
	if premiumNotificationEmail == "" {
		premiumNotificationEmail = "your-second-email@example.com"
	}

	createProductTopic.AddSubscription(awssnssubscriptions.NewEmailSubscription(_jsii_.String(budgetNotificationEmail), &awssnssubscriptions.EmailSubscriptionProps{
		FilterPolicy: &map[string]awssns.SubscriptionFilter{
			"priceCategory": awssns.SubscriptionFilter_StringFilter(&awssns.StringConditions{
				Allowlist: &[]*string{_jsii_.String("budget")},
			}),
		},
	}))

	createProductTopic.AddSubscription(awssnssubscriptions.NewEmailSubscription(_jsii_.String(premiumNotificationEmail), &awssnssubscriptions.EmailSubscriptionProps{
		FilterPolicy: &map[string]awssns.SubscriptionFilter{
			"priceCategory": awssns.SubscriptionFilter_StringFilter(&awssns.StringConditions{
				Allowlist: &[]*string{_jsii_.String("premium")},
			}),
		},
	}))

	catalogBatchProcessFn := awslambda.NewFunction(stack, _jsii_.String("catalog-batch-process-lambda"), &awslambda.FunctionProps{
		FunctionName: _jsii_.String("catalogBatchProcess"),
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      _jsii_.String("bootstrap"),
		Code:         awslambda.Code_FromAsset(lambdaAssetPath("catalog-batch-process"), nil),
		MemorySize:   _jsii_.Number(512),
		Timeout:      awscdk.Duration_Seconds(_jsii_.Number(10)),
		Environment: &map[string]*string{
			"PRODUCTS_TABLE_NAME":      productsTable.TableName(),
			"STOCKS_TABLE_NAME":        stocksTable.TableName(),
			"CREATE_PRODUCT_TOPIC_ARN": createProductTopic.TopicArn(),
		},
	})

	importFileParserFn.AddEnvironment(_jsii_.String("CATALOG_ITEMS_QUEUE_URL"), catalogItemsQueue.QueueUrl(), nil)

	catalogBatchProcessFn.AddEventSource(awslambdaeventsources.NewSqsEventSource(catalogItemsQueue, &awslambdaeventsources.SqsEventSourceProps{
		BatchSize: _jsii_.Number(5),
	}))

	productsTable.GrantReadData(listProductsFn)
	stocksTable.GrantReadData(listProductsFn)
	productsTable.GrantReadData(getProductByIDFn)
	stocksTable.GrantReadData(getProductByIDFn)
	productsTable.GrantWriteData(createProductFn)
	stocksTable.GrantWriteData(createProductFn)
	productsTable.GrantWriteData(catalogBatchProcessFn)
	stocksTable.GrantWriteData(catalogBatchProcessFn)
	catalogItemsQueue.GrantConsumeMessages(catalogBatchProcessFn)
	catalogItemsQueue.GrantSendMessages(importFileParserFn)
	createProductTopic.GrantPublish(catalogBatchProcessFn)
	importsBucket.GrantPut(importProductsFileFn, _jsii_.String("uploaded/*"))
	importsBucket.GrantRead(importFileParserFn, _jsii_.String("uploaded/*"))
	importsBucket.GrantDelete(importFileParserFn, _jsii_.String("uploaded/*"))
	importsBucket.GrantPut(importFileParserFn, _jsii_.String("parsed/*"))

	importsBucket.AddEventNotification(
		awss3.EventType_OBJECT_CREATED,
		awss3notifications.NewLambdaDestination(importFileParserFn),
		&awss3.NotificationKeyFilter{Prefix: _jsii_.String("uploaded/")},
	)

	api := awsapigateway.NewRestApi(stack, _jsii_.String("product-service-api"), &awsapigateway.RestApiProps{
		RestApiName: _jsii_.String("product-service"),
		Description: _jsii_.String("REST API for product service handlers"),
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowMethods: &[]*string{
				_jsii_.String("GET"),
				_jsii_.String("POST"),
				_jsii_.String("OPTIONS"),
			},
			AllowOrigins: &[]*string{allowedOrigin},
			AllowHeaders: &[]*string{
				_jsii_.String("Content-Type"),
				_jsii_.String("Authorization"),
				_jsii_.String("X-Amz-Date"),
				_jsii_.String("X-Api-Key"),
				_jsii_.String("X-Amz-Security-Token"),
			},
			AllowCredentials: _jsii_.Bool(true),
		},
		DeployOptions: &awsapigateway.StageOptions{
			StageName: _jsii_.String("prod"),
		},
	})

	existingUserPoolID := existingCognitoUserPoolID()
	existingUserPoolClientID := existingCognitoUserPoolClientID()

	if existingUserPoolID == "" && existingUserPoolClientID != "" {
		panic("EXISTING_COGNITO_USER_POOL_CLIENT_ID requires EXISTING_COGNITO_USER_POOL_ID")
	}

	if existingUserPoolID != "" && existingUserPoolClientID == "" {
		panic("EXISTING_COGNITO_USER_POOL_ID requires EXISTING_COGNITO_USER_POOL_CLIENT_ID")
	}

	var productUserPool awscognito.IUserPool
	var productUserPoolID *string
	var productUserPoolClientID *string

	if existingUserPoolID != "" {
		productUserPool = awscognito.UserPool_FromUserPoolId(stack, _jsii_.String("products-user-pool-import"), _jsii_.String(existingUserPoolID))
		productUserPoolID = _jsii_.String(existingUserPoolID)
		productUserPoolClientID = _jsii_.String(existingUserPoolClientID)
	} else {
		createdProductUserPool := awscognito.NewUserPool(stack, _jsii_.String("products-user-pool"), &awscognito.UserPoolProps{
			UserPoolName:      _jsii_.String("products-user-pool"),
			SelfSignUpEnabled: _jsii_.Bool(false),
			SignInAliases: &awscognito.SignInAliases{
				Email: _jsii_.Bool(true),
			},
			StandardAttributes: &awscognito.StandardAttributes{
				Email: &awscognito.StandardAttribute{
					Required: _jsii_.Bool(true),
					Mutable:  _jsii_.Bool(true),
				},
			},
			PasswordPolicy: &awscognito.PasswordPolicy{
				MinLength: _jsii_.Number(8),
			},
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		})

		createdProductUserPoolClient := createdProductUserPool.AddClient(_jsii_.String("products-user-pool-client"), &awscognito.UserPoolClientOptions{
			UserPoolClientName: _jsii_.String("products-user-pool-client"),
			AuthFlows: &awscognito.AuthFlow{
				UserPassword: _jsii_.Bool(true),
				UserSrp:      _jsii_.Bool(true),
			},
		})

		productUserPool = createdProductUserPool
		productUserPoolID = createdProductUserPool.UserPoolId()
		productUserPoolClientID = createdProductUserPoolClient.UserPoolClientId()
	}

	cognitoAuthorizer := awsapigateway.NewCognitoUserPoolsAuthorizer(stack, _jsii_.String("products-cognito-authorizer"), &awsapigateway.CognitoUserPoolsAuthorizerProps{
		AuthorizerName: _jsii_.String("products-cognito-authorizer"),
		CognitoUserPools: &[]awscognito.IUserPool{
			productUserPool,
		},
		IdentitySource: _jsii_.String("method.request.header.Authorization"),
	})

	basicAuthorizerFn := awslambda.Function_FromFunctionArn(stack, _jsii_.String("basic-authorizer-import"), basicAuthorizerArn)

	basicAuthorizer := awsapigateway.NewTokenAuthorizer(stack, _jsii_.String("basic-authorizer"), &awsapigateway.TokenAuthorizerProps{
		AuthorizerName:  _jsii_.String("basic-authorizer"),
		Handler:         basicAuthorizerFn,
		IdentitySource:  _jsii_.String("method.request.header.Authorization"),
		ResultsCacheTtl: awscdk.Duration_Seconds(_jsii_.Number(0)),
	})

	productsResource := api.Root().AddResource(_jsii_.String("products"), nil)
	productsResource.AddMethod(
		_jsii_.String("GET"),
		awsapigateway.NewLambdaIntegration(listProductsFn, nil),
		&awsapigateway.MethodOptions{
			AuthorizationType: awsapigateway.AuthorizationType_COGNITO,
			Authorizer:        cognitoAuthorizer,
		},
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

	importResource := api.Root().AddResource(_jsii_.String("import"), nil)
	importResource.AddMethod(
		_jsii_.String("GET"),
		awsapigateway.NewLambdaIntegration(importProductsFileFn, nil),
		&awsapigateway.MethodOptions{
			AuthorizationType: awsapigateway.AuthorizationType_CUSTOM,
			Authorizer:        basicAuthorizer,
			RequestParameters: &map[string]*bool{
				"method.request.querystring.name": _jsii_.Bool(true),
			},
		},
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

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-product-json-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/product-service/swagger.json"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-product-yaml-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/product-service/swagger.yaml"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-import-json-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/import-service/swagger.json"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-import-yaml-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/import-service/swagger.yaml"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-authorization-json-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/authorization-service/swagger.json"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("swagger-authorization-yaml-url"), &awscdk.CfnOutputProps{
		Value: awscdk.Fn_Join(_jsii_.String(""), &[]*string{
			_jsii_.String("https://"),
			distribution.DistributionDomainName(),
			_jsii_.String("/docs/authorization-service/swagger.yaml"),
		}),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("get-products-list-lambda-name"), &awscdk.CfnOutputProps{
		Value: listProductsFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("products-user-pool-id"), &awscdk.CfnOutputProps{
		Value: productUserPoolID,
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("products-user-pool-client-id"), &awscdk.CfnOutputProps{
		Value: productUserPoolClientID,
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("get-product-by-id-lambda-name"), &awscdk.CfnOutputProps{
		Value: getProductByIDFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("create-product-lambda-name"), &awscdk.CfnOutputProps{
		Value: createProductFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("import-products-file-lambda-name"), &awscdk.CfnOutputProps{
		Value: importProductsFileFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("import-file-parser-lambda-name"), &awscdk.CfnOutputProps{
		Value: importFileParserFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("catalog-batch-process-lambda-name"), &awscdk.CfnOutputProps{
		Value: catalogBatchProcessFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("basic-authorizer-lambda-name"), &awscdk.CfnOutputProps{
		Value: basicAuthorizerFn.FunctionName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("imports-bucket-name"), &awscdk.CfnOutputProps{
		Value: importsBucket.BucketName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("products-table-name"), &awscdk.CfnOutputProps{
		Value: productsTable.TableName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("stocks-table-name"), &awscdk.CfnOutputProps{
		Value: stocksTable.TableName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("catalog-items-queue-name"), &awscdk.CfnOutputProps{
		Value: catalogItemsQueue.QueueName(),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("create-product-topic-arn"), &awscdk.CfnOutputProps{
		Value: createProductTopic.TopicArn(),
	})

	return stack
}
