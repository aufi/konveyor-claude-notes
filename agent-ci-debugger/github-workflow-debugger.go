package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// WorkflowRun represents a GitHub Actions workflow run
type WorkflowRun struct {
	URL          string
	RunID        string
	Repository   string
	Status       string
	Conclusion   string
	FailedLogs   string
	FullLogs     string
	ErrorSummary ErrorSummary
}

// ErrorSummary contains structured information about the failure
type ErrorSummary struct {
	FailedJobs    []string
	ErrorMessages []string
	Timeouts      []string
	FailedTests   []string
	StackTraces   []string
	ExitCodes     []int
}

// FixProposal represents a proposed fix for the workflow failure
type FixProposal struct {
	RootCause    string
	Analysis     string
	ProposedFix  string
	FilesToCheck []string
	CodeChanges  []CodeChange
	Confidence   string
}

// CodeChange represents a suggested code modification
type CodeChange struct {
	File        string
	Description string
	DiffSnippet string
}

// GitHubWorkflowDebugger is the main AI agent
type GitHubWorkflowDebugger struct {
	openaiClient *openai.Client
	apiKey       string
	model        string
}

// NewGitHubWorkflowDebugger creates a new debugger agent
func NewGitHubWorkflowDebugger(apiKey string) *GitHubWorkflowDebugger {
	client := openai.NewClient(apiKey)

	// Check for model override from environment
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = openai.GPT4oMini // Default: GPT-4o-mini for cost efficiency
		// Alternative models:
		// openai.GPT4o           - Better quality, higher cost
		// openai.GPT4Turbo       - GPT-4 Turbo
		// "gpt-4o-mini"          - Use string directly for newer models
	}

	return &GitHubWorkflowDebugger{
		openaiClient: client,
		apiKey:       apiKey,
		model:        model,
	}
}

// ParseWorkflowURL extracts repository, run ID, and optional job ID from GitHub Actions URL
// Supports both formats:
// - https://github.com/{owner}/{repo}/actions/runs/{run_id}
// - https://github.com/{owner}/{repo}/actions/runs/{run_id}/job/{job_id}
func ParseWorkflowURL(url string) (repo, runID, jobID string, err error) {
	log.Printf("Parsing URL: %s", url)

	// Try job URL format first (more specific)
	jobRe := regexp.MustCompile(`github\.com/([^/]+/[^/]+)/actions/runs/(\d+)/job/(\d+)`)
	matches := jobRe.FindStringSubmatch(url)

	if len(matches) == 4 {
		log.Printf("Detected job URL - Repo: %s, Run ID: %s, Job ID: %s", matches[1], matches[2], matches[3])
		return matches[1], matches[2], matches[3], nil
	}

	// Try workflow run URL format
	runRe := regexp.MustCompile(`github\.com/([^/]+/[^/]+)/actions/runs/(\d+)`)
	matches = runRe.FindStringSubmatch(url)

	if len(matches) == 3 {
		log.Printf("Detected workflow URL - Repo: %s, Run ID: %s", matches[1], matches[2])
		return matches[1], matches[2], "", nil
	}

	log.Printf("ERROR: URL format not recognized")
	return "", "", "", fmt.Errorf("invalid GitHub Actions URL format (expected workflow or job URL)")
}

// FetchWorkflowData retrieves workflow run data using GitHub CLI
func (d *GitHubWorkflowDebugger) FetchWorkflowData(workflowURL string) (*WorkflowRun, error) {
	log.Printf("Starting workflow data fetch...")

	repo, runID, jobID, err := ParseWorkflowURL(workflowURL)
	if err != nil {
		return nil, err
	}

	run := &WorkflowRun{
		URL:        workflowURL,
		RunID:      runID,
		Repository: repo,
	}

	log.Printf("Fetching workflow status for run %s in repo %s...", runID, repo)

	// Get workflow run status
	statusCmd := exec.Command("gh", "run", "view", runID, "--repo", repo, "--json", "status,conclusion")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	var statusData map[string]string
	if err := json.Unmarshal(statusOutput, &statusData); err != nil {
		return nil, fmt.Errorf("failed to parse status: %w", err)
	}

	run.Status = statusData["status"]
	run.Conclusion = statusData["conclusion"]

	log.Printf("Workflow status: %s, conclusion: %s", run.Status, run.Conclusion)

	// Get logs - either for specific job or all failed jobs
	var failedLogsOutput []byte
	if jobID != "" {
		// Fetch logs for specific job
		log.Printf("Fetching logs for specific job: %s", jobID)
		jobLogsCmd := exec.Command("gh", "run", "view", runID, "--repo", repo, "--log", "--job", jobID)
		failedLogsOutput, err = jobLogsCmd.Output()
		if err != nil {
			log.Printf("Warning: failed to get job logs: %v", err)
			log.Printf("Falling back to all failed logs...")
			// Fallback to failed logs
			failedLogsCmd := exec.Command("gh", "run", "view", runID, "--repo", repo, "--log-failed")
			failedLogsOutput, _ = failedLogsCmd.Output()
		} else {
			log.Printf("Successfully fetched job logs (%d bytes)", len(failedLogsOutput))
		}
	} else {
		// Get all failed job logs
		log.Printf("Fetching all failed job logs...")
		failedLogsCmd := exec.Command("gh", "run", "view", runID, "--repo", repo, "--log-failed")
		failedLogsOutput, err = failedLogsCmd.Output()
		if err != nil {
			log.Printf("Warning: failed to get failed logs: %v", err)
		} else {
			log.Printf("Successfully fetched failed logs (%d bytes)", len(failedLogsOutput))
		}
	}

	run.FailedLogs = string(failedLogsOutput)

	// Parse error summary
	log.Printf("Parsing error summary from logs...")
	run.ErrorSummary = d.parseErrorSummary(run.FailedLogs)
	log.Printf("Found %d failed jobs, %d error messages, %d timeouts, %d failed tests",
		len(run.ErrorSummary.FailedJobs),
		len(run.ErrorSummary.ErrorMessages),
		len(run.ErrorSummary.Timeouts),
		len(run.ErrorSummary.FailedTests))

	return run, nil
}

// parseErrorSummary extracts structured error information from logs
func (d *GitHubWorkflowDebugger) parseErrorSummary(logs string) ErrorSummary {
	summary := ErrorSummary{
		FailedJobs:    []string{},
		ErrorMessages: []string{},
		Timeouts:      []string{},
		FailedTests:   []string{},
		StackTraces:   []string{},
		ExitCodes:     []int{},
	}

	lines := strings.Split(logs, "\n")

	// Extract job names
	jobRe := regexp.MustCompile(`^([^/]+) / ([^/]+)\s+`)
	seenJobs := make(map[string]bool)

	// Extract error patterns
	for _, line := range lines {
		// Job names
		if matches := jobRe.FindStringSubmatch(line); len(matches) > 2 {
			job := matches[1] + " / " + matches[2]
			if !seenJobs[job] {
				summary.FailedJobs = append(summary.FailedJobs, job)
				seenJobs[job] = true
			}
		}

		// Timeout messages
		if strings.Contains(line, "Timed out") || strings.Contains(line, "timeout") {
			summary.Timeouts = append(summary.Timeouts, strings.TrimSpace(line))
		}

		// Error messages
		if strings.Contains(line, "Error:") || strings.Contains(line, "ERROR") {
			summary.ErrorMessages = append(summary.ErrorMessages, strings.TrimSpace(line))
		}

		// Test failures
		if strings.Contains(line, ".go:") && (strings.Contains(line, "FAIL") || strings.Contains(line, "Error")) {
			summary.FailedTests = append(summary.FailedTests, strings.TrimSpace(line))
		}

		// Exit codes
		exitCodeRe := regexp.MustCompile(`exit code (\d+)`)
		if matches := exitCodeRe.FindStringSubmatch(line); len(matches) > 1 {
			var code int
			fmt.Sscanf(matches[1], "%d", &code)
			summary.ExitCodes = append(summary.ExitCodes, code)
		}
	}

	return summary
}

// AnalyzeFailure uses OpenAI to analyze the workflow failure
func (d *GitHubWorkflowDebugger) AnalyzeFailure(ctx context.Context, run *WorkflowRun) (*FixProposal, error) {
	log.Printf("Building analysis prompt...")

	// Build analysis prompt
	prompt := d.buildAnalysisPrompt(run)

	promptTokens := estimateTokens(prompt)
	log.Printf("Prompt size: %d characters, estimated %d tokens", len(prompt), promptTokens)
	log.Printf("Using AI model: %s", d.model)

	// Call OpenAI API
	log.Printf("Calling OpenAI API...")
	resp, err := d.openaiClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: d.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert DevOps engineer specializing in debugging CI/CD workflows and GitHub Actions failures. You provide detailed, actionable analysis and fixes.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   8000,
			Temperature: 0.7,
		},
	)

	if err != nil {
		log.Printf("ERROR: OpenAI API call failed: %v", err)
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(resp.Choices) == 0 {
		log.Printf("ERROR: No response from OpenAI API")
		return nil, fmt.Errorf("no response from OpenAI API")
	}

	responseText := resp.Choices[0].Message.Content
	log.Printf("Received AI response (%d characters)", len(responseText))
	log.Printf("API usage - Prompt tokens: %d, Completion tokens: %d, Total: %d",
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens)

	// Parse the response into a structured fix proposal
	log.Printf("Parsing fix proposal from AI response...")
	proposal := d.parseFixProposal(responseText, run)

	return proposal, nil
}

// estimateTokens estimates the number of tokens in a string
// Conservative approximation: 1 token ~= 2.5 characters for code/logs
// (English prose is ~4 chars/token, but logs/code are denser)
func estimateTokens(text string) int {
	return int(float64(len(text)) / 2.5)
}

// filterRelevantLogs extracts the most relevant parts of logs
func (d *GitHubWorkflowDebugger) filterRelevantLogs(logs string, maxChars int) string {
	log.Printf("Filtering logs - input: %d chars, max: %d chars", len(logs), maxChars)

	lines := strings.Split(logs, "\n")

	// Priority keywords that indicate important information
	errorKeywords := []string{
		"error", "ERROR", "Error:", "failed", "FAILED", "Failed",
		"fatal", "FATAL", "panic", "PANIC",
		"Timed out", "timeout", "TIMEOUT",
		"exit code", "Exit code",
		"assertion", "expected", "actual",
		"stack trace", "Stack trace",
		"FAIL:", "FAIL ", "✗", "❌",
	}

	var relevantLines []string
	var normalLines []string

	// Separate high-priority lines from normal lines
	for _, line := range lines {
		isRelevant := false
		lineLower := strings.ToLower(line)

		for _, keyword := range errorKeywords {
			if strings.Contains(lineLower, strings.ToLower(keyword)) {
				isRelevant = true
				break
			}
		}

		if isRelevant {
			relevantLines = append(relevantLines, line)
		} else {
			normalLines = append(normalLines, line)
		}
	}

	var result strings.Builder
	currentSize := 0

	// Add all relevant lines first
	for _, line := range relevantLines {
		if currentSize+len(line)+1 > maxChars {
			break
		}
		result.WriteString(line)
		result.WriteString("\n")
		currentSize += len(line) + 1
	}

	log.Printf("Added %d relevant/error lines (%d chars)", len(relevantLines), currentSize)

	// Add context from end of logs (usually contains the actual failure)
	remainingChars := maxChars - currentSize
	if remainingChars > 0 && len(normalLines) > 0 {
		// Take the last portion of normal lines
		allNormalText := strings.Join(normalLines, "\n")
		if len(allNormalText) > remainingChars {
			result.WriteString("\n...[middle section omitted]...\n\n")
			truncatedPart := allNormalText[len(allNormalText)-remainingChars:]
			result.WriteString(truncatedPart)
			log.Printf("Added %d chars from end of logs (truncated from %d chars)", len(truncatedPart), len(allNormalText))
		} else {
			result.WriteString(allNormalText)
			log.Printf("Added all %d normal lines (%d chars)", len(normalLines), len(allNormalText))
		}
	}

	filteredResult := result.String()
	log.Printf("Log filtering complete - output: %d chars (%.1f%% of input)",
		len(filteredResult),
		float64(len(filteredResult))/float64(len(logs))*100)

	return filteredResult
}

// buildAnalysisPrompt creates the prompt for the AI
func (d *GitHubWorkflowDebugger) buildAnalysisPrompt(run *WorkflowRun) string {
	var sb strings.Builder

	sb.WriteString("Analyze this GitHub Actions workflow failure and provide a comprehensive diagnosis.\n\n")
	sb.WriteString(fmt.Sprintf("## Workflow Information\n"))
	sb.WriteString(fmt.Sprintf("- URL: %s\n", run.URL))
	sb.WriteString(fmt.Sprintf("- Repository: %s\n", run.Repository))
	sb.WriteString(fmt.Sprintf("- Run ID: %s\n", run.RunID))
	sb.WriteString(fmt.Sprintf("- Status: %s\n", run.Status))
	sb.WriteString(fmt.Sprintf("- Conclusion: %s\n\n", run.Conclusion))

	sb.WriteString("## Error Summary\n")
	if len(run.ErrorSummary.FailedJobs) > 0 {
		sb.WriteString(fmt.Sprintf("Failed Jobs (%d total):\n", len(run.ErrorSummary.FailedJobs)))
		// Show only first 5 job names, not the full details
		for i, job := range run.ErrorSummary.FailedJobs {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(run.ErrorSummary.FailedJobs)-5))
				break
			}
			sb.WriteString(fmt.Sprintf("  - %s\n", job))
		}
	}

	// Skip detailed lists - just show counts
	if len(run.ErrorSummary.Timeouts) > 0 {
		sb.WriteString(fmt.Sprintf("Timeout messages: %d\n", len(run.ErrorSummary.Timeouts)))
	}
	if len(run.ErrorSummary.FailedTests) > 0 {
		sb.WriteString(fmt.Sprintf("Failed tests: %d\n", len(run.ErrorSummary.FailedTests)))
	}
	if len(run.ErrorSummary.ExitCodes) > 0 {
		uniqueCodes := make(map[int]bool)
		for _, code := range run.ErrorSummary.ExitCodes {
			uniqueCodes[code] = true
		}
		codes := []int{}
		for code := range uniqueCodes {
			codes = append(codes, code)
		}
		sb.WriteString(fmt.Sprintf("Exit Codes: %v\n", codes))
	}

	// Calculate how much space we have for logs
	// OpenAI limit: 128k tokens total
	// Reserve for response: 8k tokens (MaxTokens setting)
	// Available for input: 120k tokens
	// 1 token ~= 2.5 chars for logs/code (conservative)
	//
	// Token budget breakdown:
	// - System message: ~50 tokens
	// - Workflow info: ~50 tokens
	// - Error summary: ~500 tokens
	// - Task instructions: ~200 tokens
	// - Log overhead (formatting): ~200 tokens
	// - Total overhead: ~1000 tokens
	//
	// Safe budget for actual logs: 118k - 1k = 117k tokens
	// 117k tokens * 2.5 chars/token = ~292k chars
	// Be very conservative: use 30k chars (~12k tokens) to ensure we stay safe
	maxLogChars := 30000

	currentPromptSize := sb.Len()
	remainingChars := maxLogChars - currentPromptSize

	sb.WriteString("\n## Failed Job Logs\n")
	sb.WriteString("```\n")

	filteredLogs := d.filterRelevantLogs(run.FailedLogs, remainingChars)
	sb.WriteString(filteredLogs)

	sb.WriteString("\n```\n\n")

	sb.WriteString("## Task\n")
	sb.WriteString("Please analyze this workflow failure and provide:\n\n")
	sb.WriteString("1. **Root Cause**: What is the fundamental issue causing the failure?\n")
	sb.WriteString("2. **Detailed Analysis**: Explain what went wrong, including:\n")
	sb.WriteString("   - Which component/test failed\n")
	sb.WriteString("   - Why it failed (timeout, assertion, error, etc.)\n")
	sb.WriteString("   - Any relevant context from the logs\n")
	sb.WriteString("3. **Proposed Fix**: Specific, actionable steps to resolve the issue\n")
	sb.WriteString("4. **Files to Check**: Which files should be examined or modified\n")
	sb.WriteString("5. **Code Changes**: If applicable, suggest specific code modifications\n")
	sb.WriteString("6. **Confidence Level**: Rate your confidence in this diagnosis (High/Medium/Low)\n\n")
	sb.WriteString("Format your response with clear markdown sections using the headers above.\n")

	finalPrompt := sb.String()
	estimatedTokens := estimateTokens(finalPrompt)

	// Log token estimate for debugging
	log.Printf("Estimated tokens: %d (max: 128000)", estimatedTokens)

	return finalPrompt
}

// parseFixProposal extracts structured information from AI response
func (d *GitHubWorkflowDebugger) parseFixProposal(response string, run *WorkflowRun) *FixProposal {
	proposal := &FixProposal{}

	// Extract sections using regex
	rootCauseRe := regexp.MustCompile(`(?i)##?\s*Root Cause[:\s]*\n((?s:.*?))(?:\n##|\n\n##|\z)`)
	analysisRe := regexp.MustCompile(`(?i)##?\s*Detailed Analysis[:\s]*\n((?s:.*?))(?:\n##|\n\n##|\z)`)
	fixRe := regexp.MustCompile(`(?i)##?\s*Proposed Fix[:\s]*\n((?s:.*?))(?:\n##|\n\n##|\z)`)
	filesRe := regexp.MustCompile(`(?i)##?\s*Files to Check[:\s]*\n((?s:.*?))(?:\n##|\n\n##|\z)`)
	confidenceRe := regexp.MustCompile(`(?i)##?\s*Confidence Level[:\s]*\n?\s*([^\n]+)`)

	if matches := rootCauseRe.FindStringSubmatch(response); len(matches) > 1 {
		proposal.RootCause = strings.TrimSpace(matches[1])
	}

	if matches := analysisRe.FindStringSubmatch(response); len(matches) > 1 {
		proposal.Analysis = strings.TrimSpace(matches[1])
	}

	if matches := fixRe.FindStringSubmatch(response); len(matches) > 1 {
		proposal.ProposedFix = strings.TrimSpace(matches[1])
	}

	if matches := filesRe.FindStringSubmatch(response); len(matches) > 1 {
		filesText := strings.TrimSpace(matches[1])
		// Extract file paths (look for lines starting with - or containing .go, .yaml, etc.)
		fileLines := strings.Split(filesText, "\n")
		for _, line := range fileLines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
				file := strings.TrimPrefix(strings.TrimPrefix(line, "-"), "*")
				file = strings.TrimSpace(file)
				if file != "" {
					proposal.FilesToCheck = append(proposal.FilesToCheck, file)
				}
			}
		}
	}

	if matches := confidenceRe.FindStringSubmatch(response); len(matches) > 1 {
		proposal.Confidence = strings.TrimSpace(matches[1])
	}

	return proposal
}

// GenerateReport creates a formatted report of the analysis
func (d *GitHubWorkflowDebugger) GenerateReport(run *WorkflowRun, proposal *FixProposal) string {
	var sb strings.Builder

	sb.WriteString("# GitHub Workflow Failure Analysis Report\n\n")
	sb.WriteString(fmt.Sprintf("**Workflow URL**: %s\n", run.URL))
	sb.WriteString(fmt.Sprintf("**Repository**: %s\n", run.Repository))
	sb.WriteString(fmt.Sprintf("**Run ID**: %s\n", run.RunID))
	sb.WriteString(fmt.Sprintf("**Conclusion**: %s\n\n", run.Conclusion))

	sb.WriteString("---\n\n")

	sb.WriteString("## Root Cause\n\n")
	sb.WriteString(proposal.RootCause)
	sb.WriteString("\n\n")

	sb.WriteString("## Detailed Analysis\n\n")
	sb.WriteString(proposal.Analysis)
	sb.WriteString("\n\n")

	sb.WriteString("## Proposed Fix\n\n")
	sb.WriteString(proposal.ProposedFix)
	sb.WriteString("\n\n")

	if len(proposal.FilesToCheck) > 0 {
		sb.WriteString("## Files to Check\n\n")
		for _, file := range proposal.FilesToCheck {
			sb.WriteString(fmt.Sprintf("- %s\n", file))
		}
		sb.WriteString("\n")
	}

	if len(proposal.CodeChanges) > 0 {
		sb.WriteString("## Suggested Code Changes\n\n")
		for i, change := range proposal.CodeChanges {
			sb.WriteString(fmt.Sprintf("### Change %d: %s\n\n", i+1, change.File))
			sb.WriteString(fmt.Sprintf("%s\n\n", change.Description))
			if change.DiffSnippet != "" {
				sb.WriteString("```diff\n")
				sb.WriteString(change.DiffSnippet)
				sb.WriteString("\n```\n\n")
			}
		}
	}

	sb.WriteString(fmt.Sprintf("**Confidence Level**: %s\n\n", proposal.Confidence))

	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("*AI Model: %s*\n", d.model))
	sb.WriteString(fmt.Sprintf("*Generated at %s*\n", time.Now().Format(time.RFC3339)))

	return sb.String()
}

// Debug is the main entry point for the agent
func (d *GitHubWorkflowDebugger) Debug(ctx context.Context, workflowURL string) (string, error) {
	log.Printf("=== GitHub Workflow Debugger Started ===")
	log.Printf("Workflow URL: %s", workflowURL)

	fmt.Println("Fetching workflow data...")
	run, err := d.FetchWorkflowData(workflowURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch workflow data: %w", err)
	}

	fmt.Printf("Workflow Status: %s (%s)\n", run.Status, run.Conclusion)
	log.Printf("Workflow data fetched successfully")

	fmt.Printf("Analyzing failure with AI...\n")

	proposal, err := d.AnalyzeFailure(ctx, run)
	if err != nil {
		return "", fmt.Errorf("failed to analyze failure: %w", err)
	}

	log.Printf("AI analysis completed successfully")
	log.Printf("Generating final report...")

	report := d.GenerateReport(run, proposal)

	log.Printf("Report generated (%d characters)", len(report))
	log.Printf("=== GitHub Workflow Debugger Completed Successfully ===")

	return report, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: github-workflow-debugger <workflow-or-job-url>")
		fmt.Println("Examples:")
		fmt.Println("  Workflow: github-workflow-debugger https://github.com/konveyor/ci/actions/runs/19353355807")
		fmt.Println("  Job:      github-workflow-debugger https://github.com/konveyor/ci/actions/runs/19353355807/job/55364349255")
		os.Exit(1)
	}

	workflowURL := os.Args[1]

	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	log.Printf("Initializing debugger...")

	// Create debugger
	debugger := NewGitHubWorkflowDebugger(apiKey)

	modelUsed := os.Getenv("OPENAI_MODEL")
	if modelUsed == "" {
		modelUsed = "gpt-4o-mini (default)"
	}
	log.Printf("AI Model: %s", modelUsed)

	// Run analysis
	ctx := context.Background()
	report, err := debugger.Debug(ctx, workflowURL)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Print report
	fmt.Println("\n" + report)

	// Save report to file
	reportFile := fmt.Sprintf("workflow-debug-%s.md", time.Now().Format("20060102-150405"))
	log.Printf("Saving report to: %s", reportFile)
	if err := os.WriteFile(reportFile, []byte(report), 0644); err != nil {
		log.Printf("Warning: failed to save report to file: %v", err)
	} else {
		fmt.Printf("\nReport saved to: %s\n", reportFile)
		log.Printf("Report file saved successfully")
	}
}
