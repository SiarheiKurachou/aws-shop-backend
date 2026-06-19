package stack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticbeanstalk"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	_jsii_ "github.com/aws/jsii-runtime-go"
)

func bffSourceAssetPath() *string {
	candidates := []string{
		filepath.Join("dist", "bff-service-eb"),
		filepath.Join("cdk", "dist", "bff-service-eb"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return _jsii_.String(candidate)
		}
	}

	panic("BFF Elastic Beanstalk source bundle not found; run `make build-bff-eb`")
}

func requiredEnv(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		panic(fmt.Sprintf("%s is required", name))
	}

	return value
}

func optionalEnv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}

// NewBFFServiceStack deploys bff-service into Elastic Beanstalk as a single instance environment.
func NewBFFServiceStack(scope constructs.Construct) awscdk.Stack {
	githubLogin := requiredEnv("GITHUB_ACCOUNT_LOGIN")
	environmentName := optionalEnv("BFF_ENVIRONMENT_NAME", "dev")
	productURL := requiredEnv("PRODUCT_URL")
	cartURL := requiredEnv("CART_URL")
	solutionStackName := optionalEnv("BFF_EB_SOLUTION_STACK_NAME", "64bit Amazon Linux 2 v5.8.4 running Go 1")
	instanceType := optionalEnv("BFF_EB_INSTANCE_TYPE", "t3.micro")

	applicationName := fmt.Sprintf("%s-bff-api", githubLogin)
	cnamePrefix := fmt.Sprintf("%s-bff-api-%s", githubLogin, environmentName)

	stack := awscdk.NewStack(scope, _jsii_.String("bff-service-stack"), &awscdk.StackProps{
		StackName: _jsii_.String("bff-service-stack"),
		Env:       deploymentEnv(),
	})

	sourceAsset := awss3assets.NewAsset(stack, _jsii_.String("bff-service-source-asset"), &awss3assets.AssetProps{
		Path: bffSourceAssetPath(),
	})

	application := awselasticbeanstalk.NewCfnApplication(stack, _jsii_.String("bff-service-application"), &awselasticbeanstalk.CfnApplicationProps{
		ApplicationName: _jsii_.String(applicationName),
		Description:     _jsii_.String("BFF API for product and cart services"),
	})

	applicationVersion := awselasticbeanstalk.NewCfnApplicationVersion(stack, _jsii_.String("bff-service-application-version"), &awselasticbeanstalk.CfnApplicationVersionProps{
		ApplicationName: application.ApplicationName(),
		SourceBundle: &awselasticbeanstalk.CfnApplicationVersion_SourceBundleProperty{
			S3Bucket: sourceAsset.S3BucketName(),
			S3Key:    sourceAsset.S3ObjectKey(),
		},
		Description: _jsii_.String("BFF source bundle"),
	})
	applicationVersion.Node().AddDependency(application)

	environment := awselasticbeanstalk.NewCfnEnvironment(stack, _jsii_.String("bff-service-environment"), &awselasticbeanstalk.CfnEnvironmentProps{
		ApplicationName:   application.ApplicationName(),
		EnvironmentName:   _jsii_.String(cnamePrefix),
		CnamePrefix:       _jsii_.String(cnamePrefix),
		SolutionStackName: _jsii_.String(solutionStackName),
		VersionLabel:      applicationVersion.Ref(),
		OptionSettings: &[]interface{}{
			&awselasticbeanstalk.CfnEnvironment_OptionSettingProperty{
				Namespace:  _jsii_.String("aws:elasticbeanstalk:environment"),
				OptionName: _jsii_.String("EnvironmentType"),
				Value:      _jsii_.String("SingleInstance"),
			},
			&awselasticbeanstalk.CfnEnvironment_OptionSettingProperty{
				Namespace:  _jsii_.String("aws:autoscaling:launchconfiguration"),
				OptionName: _jsii_.String("InstanceType"),
				Value:      _jsii_.String(instanceType),
			},
			&awselasticbeanstalk.CfnEnvironment_OptionSettingProperty{
				Namespace:  _jsii_.String("aws:elasticbeanstalk:application:environment"),
				OptionName: _jsii_.String("PRODUCT_URL"),
				Value:      _jsii_.String(productURL),
			},
			&awselasticbeanstalk.CfnEnvironment_OptionSettingProperty{
				Namespace:  _jsii_.String("aws:elasticbeanstalk:application:environment"),
				OptionName: _jsii_.String("CART_URL"),
				Value:      _jsii_.String(cartURL),
			},
			&awselasticbeanstalk.CfnEnvironment_OptionSettingProperty{
				Namespace:  _jsii_.String("aws:elasticbeanstalk:application:environment"),
				OptionName: _jsii_.String("PORT"),
				Value:      _jsii_.String("8080"),
			},
			&awselasticbeanstalk.CfnEnvironment_OptionSettingProperty{
				Namespace:  _jsii_.String("aws:elasticbeanstalk:application:environment"),
				OptionName: _jsii_.String("BFF_PORT"),
				Value:      _jsii_.String("8080"),
			},
		},
	})
	environment.Node().AddDependency(applicationVersion)

	awscdk.NewCfnOutput(stack, _jsii_.String("bff-service-application-name"), &awscdk.CfnOutputProps{
		Value: _jsii_.String(applicationName),
	})

	awscdk.NewCfnOutput(stack, _jsii_.String("bff-service-cname-prefix"), &awscdk.CfnOutputProps{
		Value: _jsii_.String(cnamePrefix),
	})

	return stack
}
