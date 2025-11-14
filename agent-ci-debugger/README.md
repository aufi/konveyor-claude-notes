# GitHub Workflow Debugger AI Agent

An intelligent AI agent that automatically debugs GitHub Actions workflow failures and proposes fixes using OpenAI GPT-4.

## Features

- Fetches workflow run data from GitHub Actions using the GitHub CLI
- Parses and analyzes failed job logs
- Extracts structured error information (timeouts, failed tests, exit codes, stack traces)
- Uses OpenAI GPT-4o-mini to perform deep analysis of the failure (configurable to use GPT-4o for enhanced quality)
- Generates comprehensive fix proposals with:
  - Root cause analysis
  - Detailed breakdown of the failure
  - Actionable fix recommendations
  - Files that need to be checked or modified
  - Confidence rating
- Saves detailed reports in Markdown format

## Prerequisites

1. **GitHub CLI (`gh`)**: Install from https://cli.github.com/
   ```bash
   # Login to GitHub
   gh auth login
   ```

2. **OpenAI API Key**: Get your API key from https://platform.openai.com/api-keys
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   ```

3. **Go 1.21+**: Install from https://golang.org/

## Installation

```bash
cd ~/go/src/github.com/konveyor/claude-notes/agent-ci-debugger
go mod tidy
go build -o github-workflow-debugger github-workflow-debugger.go
```

## Usage

### Basic Usage

The debugger accepts either a workflow run URL or a specific job URL:

```bash
# Analyze entire workflow run (all failed jobs)
./github-workflow-debugger <workflow-url>

# Analyze specific job only
./github-workflow-debugger <job-url>
```

### Examples

**Analyze all failed jobs in a workflow:**
```bash
./github-workflow-debugger https://github.com/konveyor/ci/actions/runs/19353355807
```

**Analyze a specific failed job:**
```bash
./github-workflow-debugger https://github.com/konveyor/kantra-cli-tests/actions/runs/19351581387/job/55364349255
```

### Output

The agent will:
1. Fetch the workflow run data
2. Extract failed job logs
3. Analyze the failure using Claude AI
4. Print a detailed report to stdout
5. Save the report to a timestamped markdown file

Example output:
```
Fetching workflow data...
Workflow Status: completed (failure)
Analyzing failure with AI...

# GitHub Workflow Failure Analysis Report

**Workflow URL**: https://github.com/konveyor/ci/actions/runs/19353355807
**Repository**: konveyor/ci
**Run ID**: 19353355807
**Conclusion**: failure

---

## Root Cause

Test timeout in e2e-api-integration-tests job - the analysis task (ID 15) for the
Daytrader application exceeded the test timeout limit of approximately 1 hour.

## Detailed Analysis

The workflow failed because:
- The test in `analysis_test.go:209` timed out waiting for Task ID 15 to complete
- The task was continuously polled (every 5 seconds) but never reached completion
- The logs show repetitive GET requests to `/hub/tasks/15` returning status 200
- The test framework eventually gave up after the timeout threshold

...

**Confidence Level**: High

---

Report saved to: workflow-debug-20251114-120000.md
```

## Architecture

### Components

1. **WorkflowRun**: Represents a GitHub Actions workflow run with metadata and logs
2. **ErrorSummary**: Structured extraction of errors, timeouts, and failures
3. **FixProposal**: AI-generated analysis and fix recommendations
4. **GitHubWorkflowDebugger**: Main agent that orchestrates the workflow

### Flow

```
User Input (Workflow URL)
    ↓
Parse URL → Extract repo & run ID
    ↓
Fetch Data via GitHub CLI
    ↓
Parse Logs → Extract Error Summary
    ↓
Build Analysis Prompt
    ↓
Call Claude API
    ↓
Parse Response → Generate Fix Proposal
    ↓
Format & Save Report
```

### Key Functions

- `FetchWorkflowData()`: Retrieves workflow information using `gh` CLI
- `parseErrorSummary()`: Extracts structured error data from logs
- `AnalyzeFailure()`: Calls Claude API with comprehensive context
- `buildAnalysisPrompt()`: Creates detailed prompt for AI analysis
- `parseFixProposal()`: Structures AI response into actionable recommendations
- `GenerateReport()`: Creates markdown report

## Configuration

### Environment Variables

- `OPENAI_API_KEY` (required): Your OpenAI API key
- `OPENAI_MODEL` (optional): Override the AI model to use
- `GITHUB_TOKEN` (optional): GitHub token for private repos (set via `gh auth`)

### AI Model Selection

The agent uses **gpt-4o-mini** by default for cost efficiency. You can override this using the `OPENAI_MODEL` environment variable:

**Available Models:**
```bash
# Default (most cost-effective)
export OPENAI_MODEL="gpt-4o-mini"

# Better quality, higher cost
export OPENAI_MODEL="gpt-4o"

# GPT-4 Turbo
export OPENAI_MODEL="gpt-4-turbo"

# Use any OpenAI model by string
export OPENAI_MODEL="gpt-4"

# For future models (e.g., when GPT-5 is released)
export OPENAI_MODEL="gpt-5-mini"
```

**Example Usage:**
```bash
# Use GPT-4o for better analysis quality
export OPENAI_MODEL="gpt-4o"
./github-workflow-debugger https://github.com/org/repo/actions/runs/12345

# Or set it inline
OPENAI_MODEL="gpt-4o" ./github-workflow-debugger <url>
```

**Note:** There is currently no `gpt-5-mini` model available from OpenAI. The latest models are:
- `gpt-4o-mini` (default, best value)
- `gpt-4o` (highest quality, higher cost)
- `gpt-4-turbo` (good balance)

### Customization

You can modify the analysis prompt in `buildAnalysisPrompt()` to focus on specific aspects:

```go
sb.WriteString("Please analyze this workflow failure focusing on:\n")
sb.WriteString("1. Performance issues\n")
sb.WriteString("2. Resource constraints\n")
sb.WriteString("3. Race conditions\n")
// ... add your custom requirements
```

## Example Workflow Failures It Can Debug

- Test timeouts
- Build failures
- Deployment errors
- Flaky tests
- Resource exhaustion
- Integration test failures
- Container/pod crashes
- API errors
- Configuration issues

## Advanced Usage

### Programmatic Usage

```go
package main

import (
    "context"
    "fmt"
)

func main() {
    debugger := NewGitHubWorkflowDebugger("your-api-key")

    ctx := context.Background()
    report, err := debugger.Debug(ctx, "https://github.com/org/repo/actions/runs/123")
    if err != nil {
        panic(err)
    }

    fmt.Println(report)
}
```

### Integration with CI/CD

You can integrate this into your CI/CD pipeline to automatically debug failures:

```yaml
- name: Debug Workflow Failure
  if: failure()
  run: |
    ./github-workflow-debugger ${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}
```

## Debugging Output

The agent provides detailed debugging information to stderr while keeping user-facing output on stdout. This helps troubleshoot issues and understand the analysis process.

**Example debug output:**
```
2025/11/14 11:35:38 Initializing debugger...
2025/11/14 11:35:38 AI Model: gpt-4o-mini (default)
2025/11/14 11:35:38 === GitHub Workflow Debugger Started ===
2025/11/14 11:35:38 Parsing URL: https://github.com/org/repo/actions/runs/123/job/456
2025/11/14 11:35:38 Detected job URL - Repo: org/repo, Run ID: 123, Job ID: 456
2025/11/14 11:35:39 Workflow status: completed, conclusion: failure
2025/11/14 11:35:41 Successfully fetched job logs (135300 bytes)
2025/11/14 11:35:41 Found 991 failed jobs, 3 error messages, 1 timeouts, 0 failed tests
2025/11/14 11:35:41 Filtering logs - input: 135300 chars, max: 29131 chars
2025/11/14 11:35:41 Log filtering complete - output: 29164 chars (21.6% of input)
2025/11/14 11:35:41 Estimated tokens: 12298 (max: 128000)
2025/11/14 11:35:41 Calling OpenAI API...
2025/11/14 11:36:00 API usage - Prompt tokens: 10008, Completion tokens: 621, Total: 10629
2025/11/14 11:36:00 === GitHub Workflow Debugger Completed Successfully ===
```

**Key metrics shown:**
- URL format detection (workflow vs job)
- Log sizes (input/output/compression ratio)
- Error statistics extracted
- Token usage (estimated and actual)
- API costs (token counts)
- Processing time for each stage

**Redirect debug output to file:**
```bash
./github-workflow-debugger <url> 2> debug.log
```

For complete debugging guide including cost calculations and troubleshooting, see [DEBUGGING.md](DEBUGGING.md).

## Troubleshooting

### "Invalid GitHub Actions URL format"
- Ensure the URL follows one of these patterns:
  - Workflow: `https://github.com/{owner}/{repo}/actions/runs/{run_id}`
  - Job: `https://github.com/{owner}/{repo}/actions/runs/{run_id}/job/{job_id}`

### "failed to get workflow status"
- Make sure `gh` CLI is installed and authenticated
- Verify you have access to the repository
- Check that the run ID is correct

### "OPENAI_API_KEY environment variable is required"
- Set your API key: `export OPENAI_API_KEY="your-key"`

### "failed to call OpenAI API"
- Check your API key is valid
- Verify you have sufficient API credits
- Check your internet connection
- Check the debug output for token usage - you may be exceeding the context limit

## How It Handles Large Logs

The agent implements intelligent log filtering to stay within OpenAI's 128k token limit:

1. **Priority Filtering**: Extracts lines with error keywords (error, failed, timeout, etc.) first
2. **Smart Truncation**: Keeps the most relevant parts (errors + end of logs)
3. **Token Estimation**: Calculates approximate token usage before sending
4. **Adaptive Sizing**: Limits logs to ~80,000 characters (~20k tokens) for safety

The agent will automatically:
- Extract all error/failure messages
- Include relevant context from the end of logs
- Omit repetitive middle sections
- Show a summary count when truncating error lists

## Limitations

- Requires GitHub CLI to be installed and authenticated
- Very large workflows may have some logs omitted (prioritizes errors)
- Analysis quality depends on log verbosity and error messages
- Only analyzes failed jobs (not successful ones)
- Costs money per API call (check OpenAI pricing)
- Maximum context: ~128k tokens (~512k characters total)

## Future Enhancements

- [ ] Support for downloading artifacts for deeper analysis
- [ ] Integration with issue tracking systems (auto-create issues)
- [ ] Historical failure pattern analysis
- [ ] Automatic PR creation with fixes
- [ ] Support for other CI/CD platforms (GitLab, CircleCI, etc.)
- [ ] Web UI for easier access
- [ ] Slack/Discord notifications
- [ ] Multi-language support for error messages

## Contributing

Feel free to submit issues and enhancement requests!

## License

MIT License

## Cost Considerations

Using GPT-4o-mini (default):
- ~$0.15 per 1M input tokens
- ~$0.60 per 1M output tokens
- Typical workflow analysis: $0.01-0.05 per run

Using GPT-4o (better quality):
- ~$2.50 per 1M input tokens
- ~$10.00 per 1M output tokens
- Typical workflow analysis: $0.15-0.50 per run

See https://openai.com/api/pricing/ for current pricing.

## Author

Created with AI assistance
