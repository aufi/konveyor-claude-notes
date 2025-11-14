# Changelog

## [2.5.0] - 2025-11-14

### Added
- **Detailed Debugging Output**: Added comprehensive logging throughout the agent
  - URL parsing with detected format (workflow vs job)
  - Workflow status and conclusion
  - Log fetching progress with byte counts
  - Error summary statistics (jobs, messages, timeouts, tests)
  - Log filtering statistics (input/output sizes, compression ratio)
  - Token estimation and API usage metrics
  - AI model being used
  - Report generation progress
  - All debugging output goes to stderr, user-facing output to stdout

### Benefits
- Easier troubleshooting of failures
- Visibility into token usage and API costs
- Understanding of log filtering effectiveness
- Progress tracking for long-running analyses

## [2.4.0] - 2025-11-14

### Added
- **Job URL Support**: Can now analyze specific failed jobs directly
  - Accepts job URLs: `https://github.com/{owner}/{repo}/actions/runs/{run_id}/job/{job_id}`
  - Falls back to workflow run URL if job logs unavailable
  - Updated `ParseWorkflowURL()` to handle both URL formats
  - Updated `FetchWorkflowData()` to fetch specific job logs when job ID provided
  - Example: `./github-workflow-debugger https://github.com/org/repo/actions/runs/123/job/456`

### Changed
- Usage message now shows both workflow and job URL formats
- Documentation updated with job URL examples

## [2.3.0] - 2025-11-14

### Fixed
- **Critical**: Fixed token limit errors for very large workflows
  - Error summary was causing prompt to exceed 450k+ characters
  - Changed from showing full timeout/test lists to just counts
  - Reduced error summary from unlimited items to max 5 with counts
  - Now works with workflows that have 528k+ characters of logs
  - Prompt size reduced from 454k chars to <1k chars before logs
  - Final token count: ~12k tokens (well under 128k limit)

### Changed
- Token estimation improved: 1 token = 2.5 chars (was 4 chars)
- Log limit reduced to 30k chars (~12k tokens) for safety
- Error summary now shows counts instead of full lists
- Failed jobs limited to first 5 with count of remainder

## [2.2.0] - 2025-11-14

### Added
- **Model Configuration**: Support for `OPENAI_MODEL` environment variable
  - Override the default model without code changes
  - Supports any OpenAI model string (gpt-4o, gpt-4-turbo, gpt-4, etc.)
  - Future-proof for new models like gpt-5-mini when available
  - Example: `OPENAI_MODEL="gpt-4o" ./github-workflow-debugger <url>`

## [2.1.0] - 2025-11-14

### Fixed
- **Token Limit Error**: Fixed "maximum context length exceeded" error for large workflows
  - Implemented intelligent log filtering that prioritizes error messages
  - Added token estimation to prevent exceeding OpenAI's 128k token limit
  - Smart truncation keeps most relevant parts (errors + end of logs)
  - Increased effective log capacity to ~80,000 characters with priority filtering

### Added
- `filterRelevantLogs()`: Extracts high-priority error lines first
- `estimateTokens()`: Calculates approximate token usage
- Automatic truncation of error summary lists (shows first 10, then count)
- Debug logging showing estimated token usage

## [2.0.0] - 2025-11-14

### Changed
- **BREAKING**: Migrated from Anthropic Claude API to OpenAI API
  - Environment variable changed from `ANTHROPIC_API_KEY` to `OPENAI_API_KEY`
  - Uses GPT-4o-mini by default for cost efficiency
  - Can be configured to use GPT-4o for enhanced analysis quality
  - Base log handling increased from 15,000 to 20,000 characters

### Migration Guide

If you were using the previous version with Anthropic:

1. **Get an OpenAI API Key**:
   - Visit https://platform.openai.com/api-keys
   - Create a new API key

2. **Update Environment Variable**:
   ```bash
   # Old (no longer works):
   export ANTHROPIC_API_KEY="your-anthropic-key"

   # New (required):
   export OPENAI_API_KEY="your-openai-key"
   ```

3. **Rebuild the Agent**:
   ```bash
   cd agent-ci-debugger
   rm -f github-workflow-debugger  # Remove old binary
   go mod tidy                     # Download new dependencies
   go build -o github-workflow-debugger github-workflow-debugger.go
   ```

4. **Run as Before**:
   ```bash
   ./github-workflow-debugger https://github.com/owner/repo/actions/runs/12345
   # or
   ./example-usage.sh https://github.com/owner/repo/actions/runs/12345
   ```

### Why the Change?

- **Better Availability**: OpenAI API has higher rate limits and better availability
- **Cost Efficiency**: GPT-4o-mini offers excellent performance at lower cost
- **Flexibility**: Easy to switch between models (mini vs full) based on needs
- **Ecosystem**: Broader Go SDK support and community

### Cost Comparison

**GPT-4o-mini (default)**:
- Input: ~$0.15 per 1M tokens
- Output: ~$0.60 per 1M tokens
- Typical cost per analysis: $0.01-0.05

**GPT-4o (configurable)**:
- Input: ~$2.50 per 1M tokens
- Output: ~$10.00 per 1M tokens
- Typical cost per analysis: $0.15-0.50

See https://openai.com/api/pricing/ for current pricing.

---

## [1.0.0] - 2025-11-14

### Added
- Initial release with Anthropic Claude API
- GitHub workflow failure analysis
- Structured error extraction
- AI-powered root cause analysis
- Markdown report generation
- Example usage script
