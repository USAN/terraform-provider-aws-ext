package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	conntypes "github.com/aws/aws-sdk-go-v2/service/connect/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &AgentStatusResource{}
var _ resource.ResourceWithImportState = &AgentStatusResource{}

func NewAgentStatusResource() resource.Resource {
	return &AgentStatusResource{}
}

type AgentStatusResource struct {
	config *aws.Config
}

type AgentStatusResourceModel struct {
	Arn           types.String `tfsdk:"arn"`
	Description   types.String `tfsdk:"description"`
	AgentStatusID types.String `tfsdk:"agent_status_id"`
	InstanceID    types.String `tfsdk:"instance_id"`
	Name          types.String `tfsdk:"name"`
	State         types.String `tfsdk:"state"`
	// Tags          types.Map    `tfsdk:"tags"`
	// TagsAll       types.Map    `tfsdk:"tags_all"`
}

func (r *AgentStatusResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connect_agent_status"
}

func (r *AgentStatusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Connect agent status resource",

		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 250),
				},
			},
			"agent_status_id": schema.StringAttribute{
				Computed: true,
			},
			"instance_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 127),
				},
			},
			"state": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("ENABLED", "DISABLED"),
				},
			},
			// "tags": schema.MapAttribute{
			// 	Optional: true,
			// 	Elem:     &schema.Schema{Type: schema.TypeString},
			// },
			// "tags_all": schema.MapAttribute{
			// 	Optional: true,
			// 	Computed: true,
			// 	Elem:     &schema.Schema{Type: schema.TypeString},
			// },
		},
	}
}

func (r *AgentStatusResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*aws.Config)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *aws.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.config = config
}

func (r *AgentStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AgentStatusResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(*r.config)
	input := &connect.CreateAgentStatusInput{
		InstanceId:  aws.String(data.InstanceID.ValueString()),
		Name:        aws.String(data.Name.ValueString()),
		State:       conntypes.AgentStatusState(data.State.ValueString()),
		Description: aws.String(data.Description.ValueString()),
	}

	response, err := conn.CreateAgentStatus(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("Error creating Connect Agent Status", fmt.Sprintf("Could not create Connect Agent Status, unexpected error: %s", err))
		return
	}

	tflog.Trace(ctx, "created a resource")

	data.AgentStatusID = types.StringValue(aws.ToString(response.AgentStatusId))
	data.Arn = types.StringValue(aws.ToString(response.AgentStatusARN))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AgentStatusResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(*r.config)
	input := &connect.DescribeAgentStatusInput{
		AgentStatusId: aws.String(data.AgentStatusID.ValueString()),
		InstanceId:    aws.String(data.InstanceID.ValueString()),
	}

	response, err := conn.DescribeAgentStatus(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("Error reading Connect Agent Status", fmt.Sprintf("Could not read Connect Agent Status, unexpected error: %s", err))
		return
	}

	if response == nil || response.AgentStatus == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	data.AgentStatusID = types.StringValue(aws.ToString(response.AgentStatus.AgentStatusId))
	data.Arn = types.StringValue(aws.ToString(response.AgentStatus.AgentStatusARN))
	data.Description = types.StringValue(aws.ToString(response.AgentStatus.Description))
	data.Name = types.StringValue(aws.ToString(response.AgentStatus.Name))
	data.State = types.StringValue(string(response.AgentStatus.State))
	// data.Tags = types.MapValueFrom(context.Background(), types.StringType, response.AgentStatus.Tags)
	// data.TagsAll = types.MapValueFrom(context.Background

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AgentStatusResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := connect.NewFromConfig(*r.config)
	input := &connect.UpdateAgentStatusInput{
		AgentStatusId: aws.String(data.AgentStatusID.ValueString()),
		InstanceId:    aws.String(data.InstanceID.ValueString()),
		Name:          aws.String(data.Name.ValueString()),
		State:         conntypes.AgentStatusState(data.State.ValueString()),
		Description:   aws.String(data.Description.ValueString()),
	}

	_, err := conn.UpdateAgentStatus(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("Error updating Connect Agent Status", fmt.Sprintf("Could not update Connect Agent Status, unexpected error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AgentStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AgentStatusResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Unsupported by the API
	// conn := connect.NewFromConfig(*r.config)
	// input := &connect.DeleteAgentStatusInput{
	// 	AgentStatusId: aws.String(data.AgentStatusID.ValueString()),
	// 	InstanceId:    aws.String(data.InstanceID.ValueString()),
	// 	Name:          aws.String(data.Name.ValueString()),
	// 	State:         connect.AgentStatusState(data.State.ValueString()),
	// 	Description:   aws.String(data.Description.ValueString()),
	// }

	// _, err := conn.DeleteAgentStatus(ctx, input)

	// if err != nil {
	// 	resp.Diagnostics.AddError("Error deleting Connect Agent Status", fmt.Sprintf("Could not delete Connect Agent Status, unexpected error: %s", err))
	// 	return
	// }
}

func (r *AgentStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
