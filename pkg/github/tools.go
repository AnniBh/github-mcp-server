package github

import (
	"context"

	"github.com/github/github-mcp-server/pkg/monitoring"
	"github.com/github/github-mcp-server/pkg/raw"
	"github.com/github/github-mcp-server/pkg/toolsets"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/google/go-github/v72/github"
	"github.com/mark3labs/mcp-go/server"
	"github.com/shurcooL/githubv4"
)

type GetClientFn func(context.Context) (*github.Client, error)
type GetGQLClientFn func(context.Context) (*githubv4.Client, error)

var DefaultTools = []string{"all"}

func DefaultToolsetGroup(readOnly bool, getClient GetClientFn, getGQLClient GetGQLClientFn, getRawClient raw.GetRawClientFn, t translations.TranslationHelperFunc) *toolsets.ToolsetGroup {
	tsg := toolsets.NewToolsetGroup(readOnly)

	fetchUserDetails := func(ctx context.Context) (monitoring.UserDetails, error) {
		details, err := FetchUserDetails(ctx, getClient)
		if err != nil {
			return monitoring.UserDetails{}, err
		}
		return monitoring.UserDetails(details), nil
	}

	// Define all available features with their default state (disabled)
	// Create toolsets
	repos := toolsets.NewToolset("repos", "GitHub Repository related tools").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(SearchRepositories(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetFileContents(getClient, getRawClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListCommits(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(SearchCode(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetCommit(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListBranches(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListTags(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetTag(getClient, t)), fetchUserDetails),
		).
		AddWriteTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(CreateOrUpdateFile(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(CreateRepository(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ForkRepository(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(CreateBranch(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(PushFiles(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(DeleteFile(getClient, t)), fetchUserDetails),
		).
		AddResourceTemplates(
			toolsets.NewServerResourceTemplate(GetRepositoryResourceContent(getClient, getRawClient, t)),
			toolsets.NewServerResourceTemplate(GetRepositoryResourceBranchContent(getClient, getRawClient, t)),
			toolsets.NewServerResourceTemplate(GetRepositoryResourceCommitContent(getClient, getRawClient, t)),
			toolsets.NewServerResourceTemplate(GetRepositoryResourceTagContent(getClient, getRawClient, t)),
			toolsets.NewServerResourceTemplate(GetRepositoryResourcePrContent(getClient, getRawClient, t)),
		)
	issues := toolsets.NewToolset("issues", "GitHub Issues related tools").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(GetIssue(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(SearchIssues(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListIssues(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetIssueComments(getClient, t)), fetchUserDetails),
		).
		AddWriteTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(CreateIssue(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(AddIssueComment(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(UpdateIssue(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(AssignCopilotToIssue(getGQLClient, t)), fetchUserDetails),
		)
	users := toolsets.NewToolset("users", "GitHub User related tools").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(SearchUsers(getClient, t)), fetchUserDetails),
		)
	pullRequests := toolsets.NewToolset("pull_requests", "GitHub Pull Request related tools").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(GetPullRequest(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListPullRequests(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetPullRequestFiles(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetPullRequestStatus(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetPullRequestComments(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetPullRequestReviews(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetPullRequestDiff(getClient, t)), fetchUserDetails),
		).
		AddWriteTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(MergePullRequest(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(UpdatePullRequestBranch(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(CreatePullRequest(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(UpdatePullRequest(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(RequestCopilotReview(getClient, t)), fetchUserDetails),

			// Reviews
			monitoring.WithMonitoring(toolsets.NewServerTool(CreateAndSubmitPullRequestReview(getGQLClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(CreatePendingPullRequestReview(getGQLClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(AddPullRequestReviewCommentToPendingReview(getGQLClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(SubmitPendingPullRequestReview(getGQLClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(DeletePendingPullRequestReview(getGQLClient, t)), fetchUserDetails),
		)
	codeSecurity := toolsets.NewToolset("code_security", "Code security related tools, such as GitHub Code Scanning").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(GetCodeScanningAlert(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListCodeScanningAlerts(getClient, t)), fetchUserDetails),
		)
	secretProtection := toolsets.NewToolset("secret_protection", "Secret protection related tools, such as GitHub Secret Scanning").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(GetSecretScanningAlert(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListSecretScanningAlerts(getClient, t)), fetchUserDetails),
		)

	notifications := toolsets.NewToolset("notifications", "GitHub Notifications related tools").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(ListNotifications(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetNotificationDetails(getClient, t)), fetchUserDetails),
		).
		AddWriteTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(DismissNotification(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(MarkAllNotificationsRead(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ManageNotificationSubscription(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ManageRepositoryNotificationSubscription(getClient, t)), fetchUserDetails),
		)

	actions := toolsets.NewToolset("actions", "GitHub Actions workflows and CI/CD operations").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(ListWorkflows(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListWorkflowRuns(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetWorkflowRun(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetWorkflowRunLogs(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListWorkflowJobs(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetJobLogs(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(ListWorkflowRunArtifacts(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(DownloadWorkflowRunArtifact(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetWorkflowRunUsage(getClient, t)), fetchUserDetails),
		).
		AddWriteTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(RunWorkflow(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(RerunWorkflowRun(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(RerunFailedJobs(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(CancelWorkflowRun(getClient, t)), fetchUserDetails),
			monitoring.WithMonitoring(toolsets.NewServerTool(DeleteWorkflowRunLogs(getClient, t)), fetchUserDetails),
		)

	// Keep experiments alive so the system doesn't error out when it's always enabled
	experiments := toolsets.NewToolset("experiments", "Experimental features that are not considered stable yet")

	contextTools := toolsets.NewToolset("context", "Tools that provide context about the current user and GitHub context you are operating in").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(GetMe(getClient, t)), fetchUserDetails),
		)

	// Add toolsets to the group
	tsg.AddToolset(contextTools)
	tsg.AddToolset(repos)
	tsg.AddToolset(issues)
	tsg.AddToolset(users)
	tsg.AddToolset(pullRequests)
	tsg.AddToolset(actions)
	tsg.AddToolset(codeSecurity)
	tsg.AddToolset(secretProtection)
	tsg.AddToolset(notifications)
	tsg.AddToolset(experiments)

	return tsg
}

// InitDynamicToolset creates a dynamic toolset that can be used to enable other toolsets, and so requires the server and toolset group as arguments
func InitDynamicToolset(s *server.MCPServer, tsg *toolsets.ToolsetGroup, t translations.TranslationHelperFunc) *toolsets.Toolset {
	// Create a new dynamic toolset
	// Need to add the dynamic toolset last so it can be used to enable other toolsets
	dynamicToolSelection := toolsets.NewToolset("dynamic", "Discover GitHub MCP tools that can help achieve tasks by enabling additional sets of tools, you can control the enablement of any toolset to access its tools when this toolset is enabled.").
		AddReadTools(
			monitoring.WithMonitoring(toolsets.NewServerTool(ListAvailableToolsets(tsg, t)), nil),
			monitoring.WithMonitoring(toolsets.NewServerTool(GetToolsetsTools(tsg, t)), nil),
			monitoring.WithMonitoring(toolsets.NewServerTool(EnableToolset(s, tsg, t)), nil),
		)

	dynamicToolSelection.Enabled = true
	return dynamicToolSelection
}

// ToBoolPtr converts a bool to a *bool pointer.
func ToBoolPtr(b bool) *bool {
	return &b
}
