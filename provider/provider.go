// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure AwsExtProvider satisfies various provider interfaces.
var _ provider.Provider = &AwsExtProvider{}
var _ provider.ProviderWithFunctions = &AwsExtProvider{}
var _ provider.ProviderWithEphemeralResources = &AwsExtProvider{}

// AwsExtProvider defines the provider implementation.
type AwsExtProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AwsExtProviderModel describes the provider data model.
type AwsExtProviderModel struct {
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
	Token     types.String `tfsdk:"token"`
	Region    types.String `tfsdk:"region"`
	Profile   types.String `tfsdk:"profile"`
	RoleArn   types.String `tfsdk:"role_arn"`
}

func (p *AwsExtProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "awsext"
	resp.Version = p.version
}

func (p *AwsExtProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Description: "AWS access key",
				Optional:    true,
			},
			"secret_key": schema.StringAttribute{
				Description: "AWS secret key",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "AWS session token",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "AWS region",
				Optional:    true,
			},
			"profile": schema.StringAttribute{
				Description: "AWS profile",
				Optional:    true,
			},
			"role_arn": schema.StringAttribute{
				Description: "AWS role ARN",
				Optional:    true,
			},
		},
	}
}

func (p *AwsExtProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AwsExtProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	addendums := []func(*config.LoadOptions) error{}
	if data.AccessKey.ValueString() != "" && data.SecretKey.ValueString() != "" {
		addendums = append(addendums, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(data.AccessKey.ValueString(), data.SecretKey.ValueString(), data.Token.ValueString())))
	} else if data.Profile.ValueString() != "" {
		addendums = append(addendums, config.WithSharedConfigProfile(data.Profile.ValueString()))
	}

	if data.Region.ValueString() != "" {
		addendums = append(addendums, config.WithRegion(data.Region.ValueString()))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), addendums...)

	if err != nil {
		resp.Diagnostics.AddError("Failed to load AWS config", err.Error())
		return
	}

	if data.RoleArn.ValueString() != "" {
		stsClient := sts.NewFromConfig(cfg)
		creds := stscreds.NewAssumeRoleProvider(stsClient, data.RoleArn.ValueString())
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}

	resp.ResourceData = cfg
}

func (p *AwsExtProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAgentStatusResource,
	}
}

func (p *AwsExtProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *AwsExtProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *AwsExtProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AwsExtProvider{
			version: version,
		}
	}
}
