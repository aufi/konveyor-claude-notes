# Debugging Output Guide

The GitHub Workflow Debugger provides comprehensive debugging information to help you understand what's happening during the analysis process and troubleshoot any issues.

## How It Works

- **User-facing output** (analysis report) goes to **stdout**
- **Debugging information** goes to **stderr**
- This separation allows you to redirect debug output separately from the actual report

## Debug Output Stages

### 1. Initialization
```
2025/11/14 11:35:38 Initializing debugger...
2025/11/14 11:35:38 AI Model: gpt-4o-mini (default)
2025/11/14 11:35:38 === GitHub Workflow Debugger Started ===
```

Shows which AI model is being used (default or environment override).

### 2. URL Parsing
```
2025/11/14 11:35:38 Parsing URL: https://github.com/org/repo/actions/runs/123/job/456
2025/11/14 11:35:38 Detected job URL - Repo: org/repo, Run ID: 123, Job ID: 456
```

Or for workflow URLs:
```
2025/11/14 11:35:38 Detected workflow URL - Repo: org/repo, Run ID: 123
```

Confirms that the URL was parsed correctly and shows what type it is.

### 3. Workflow Data Fetch
```
2025/11/14 11:35:38 Starting workflow data fetch...
2025/11/14 11:35:38 Fetching workflow status for run 123 in repo org/repo...
2025/11/14 11:35:39 Workflow status: completed, conclusion: failure
```

Shows the workflow/job status retrieval process.

### 4. Log Fetching
```
2025/11/14 11:35:39 Fetching logs for specific job: 55364349255
2025/11/14 11:35:41 Successfully fetched job logs (135300 bytes)
```

Or for workflow runs:
```
2025/11/14 11:35:39 Fetching all failed job logs...
2025/11/14 11:35:41 Successfully fetched failed logs (528000 bytes)
```

Shows which logs are being fetched and their size.

### 5. Error Summary Parsing
```
2025/11/14 11:35:41 Parsing error summary from logs...
2025/11/14 11:35:41 Found 991 failed jobs, 3 error messages, 1 timeouts, 0 failed tests
```

Displays statistics about errors found in the logs.

### 6. Log Filtering
```
2025/11/14 11:35:41 Filtering logs - input: 135300 chars, max: 29131 chars
2025/11/14 11:35:41 Added 31 relevant/error lines (5520 chars)
2025/11/14 11:35:41 Added 23611 chars from end of logs (truncated from 129780 chars)
2025/11/14 11:35:41 Log filtering complete - output: 29164 chars (21.6% of input)
```

Shows how logs are filtered to fit within token limits:
- Input size
- How many error lines were extracted
- How much context was added
- Final compression ratio

### 7. AI Analysis
```
2025/11/14 11:35:41 Building analysis prompt...
2025/11/14 11:35:41 Prompt size: 30745 characters, estimated 12298 tokens
2025/11/14 11:35:41 Using AI model: gpt-4o-mini
2025/11/14 11:35:41 Calling OpenAI API...
2025/11/14 11:36:00 Received AI response (2835 characters)
2025/11/14 11:36:00 API usage - Prompt tokens: 10008, Completion tokens: 621, Total: 10629
```

Provides critical information about:
- Token estimation (before API call)
- Actual token usage (after API call)
- API cost metrics
- Response time and size

### 8. Report Generation
```
2025/11/14 11:36:00 Parsing fix proposal from AI response...
2025/11/14 11:36:00 AI analysis completed successfully
2025/11/14 11:36:00 Generating final report...
2025/11/14 11:36:00 Report generated (2196 characters)
2025/11/14 11:36:00 Saving report to: workflow-debug-20251114-113600.md
2025/11/14 11:36:00 Report file saved successfully
2025/11/14 11:36:00 === GitHub Workflow Debugger Completed Successfully ===
```

Confirms successful completion and report file location.

## Error Messages

If something goes wrong, you'll see error messages in the debug output:

```
2025/11/14 11:35:38 ERROR: URL format not recognized
2025/11/14 11:35:39 Warning: failed to get job logs: exit status 1
2025/11/14 11:35:39 Falling back to all failed logs...
2025/11/14 11:35:41 ERROR: OpenAI API call failed: status code 429
```

## Useful Commands

### Save debug output to file
```bash
./github-workflow-debugger <url> 2> debug.log
```

### Show only debug output (hide report)
```bash
./github-workflow-debugger <url> > /dev/null
```

### Save both separately
```bash
./github-workflow-debugger <url> > report.md 2> debug.log
```

### Monitor token usage
```bash
./github-workflow-debugger <url> 2>&1 | grep -i token
```

### Check API costs
```bash
./github-workflow-debugger <url> 2>&1 | grep "API usage"
```

## Understanding Token Usage

The debug output shows both **estimated** and **actual** token usage:

- **Estimated tokens**: Calculated before API call using `chars / 2.5`
- **Actual tokens**: Reported by OpenAI API after the call

**Example:**
```
Estimated tokens: 12298 (max: 128000)
API usage - Prompt tokens: 10008, Completion tokens: 621, Total: 10629
```

This shows:
- Estimated 12,298 tokens (conservative estimate)
- Actually used 10,008 prompt tokens
- Received 621 completion tokens
- Total: 10,629 tokens (well under the 128k limit)

## Cost Calculation

Use the actual token counts to calculate costs:

**With gpt-4o-mini:**
- Input: 10,008 tokens × $0.15/1M = $0.0015
- Output: 621 tokens × $0.60/1M = $0.0004
- **Total: ~$0.0019 per analysis**

**With gpt-4o:**
- Input: 10,008 tokens × $2.50/1M = $0.025
- Output: 621 tokens × $10.00/1M = $0.006
- **Total: ~$0.031 per analysis**

## Troubleshooting with Debug Output

### Problem: Token limit exceeded
Look for:
```
ERROR: OpenAI API call failed: maximum context length is 128000 tokens
```

Check the estimated tokens line - if it's close to 128000, the log filtering needs adjustment.

### Problem: Job logs not found
Look for:
```
Warning: failed to get job logs: exit status 1
Falling back to all failed logs...
```

The agent automatically falls back to workflow logs.

### Problem: No errors found
Check:
```
Found 0 failed jobs, 0 error messages, 0 timeouts, 0 failed tests
```

This might indicate the workflow wasn't actually a failure, or the log format is unexpected.

### Problem: API errors
Look for HTTP status codes:
```
ERROR: OpenAI API call failed: status code: 401
```
- 401: Invalid API key
- 429: Rate limit exceeded
- 500: OpenAI server error

## Performance Metrics

The timestamps allow you to see how long each stage takes:

```
11:35:38 - Started
11:35:39 - Got workflow status (1 second)
11:35:41 - Fetched logs (2 seconds)
11:35:41 - Filtered logs (< 1 second)
11:36:00 - Got AI response (19 seconds)
11:36:00 - Generated report (< 1 second)

Total time: ~22 seconds
```

Most of the time is spent waiting for the OpenAI API response.

## Development and Testing

When developing or testing changes to the agent, the debug output helps you:

1. **Verify URL parsing** works for both formats
2. **Check log filtering** effectiveness (compression ratio)
3. **Monitor token usage** to avoid exceeding limits
4. **Track API costs** during development
5. **Identify bottlenecks** using timestamps
6. **Catch errors** before they cause failures
