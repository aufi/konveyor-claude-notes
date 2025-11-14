# AI Model Selection Guide

## Current Available Models

As of November 2024, OpenAI provides these models:

### GPT-4o Family (Recommended)
- **gpt-4o-mini** (Default)
  - Best value for most use cases
  - Fast and cost-effective
  - ~$0.15 per 1M input tokens
  - Good for workflow debugging

- **gpt-4o**
  - Highest quality analysis
  - More comprehensive insights
  - ~$2.50 per 1M input tokens
  - Best for critical failures

### GPT-4 Turbo
- **gpt-4-turbo**
  - Good balance of quality and cost
  - Slightly older than gpt-4o
  - ~$10 per 1M input tokens

### Legacy Models
- **gpt-4**
  - Original GPT-4
  - Higher cost, slower
  - Not recommended for this use case

- **gpt-3.5-turbo**
  - Fast but lower quality
  - May miss subtle issues
  - Not recommended for debugging

## How to Change Models

### Method 1: Environment Variable (Recommended)

```bash
# Use GPT-4o for better quality
export OPENAI_MODEL="gpt-4o"
./github-workflow-debugger <url>

# Or inline
OPENAI_MODEL="gpt-4o" ./github-workflow-debugger <url>
```

### Method 2: Modify Code

Edit `github-workflow-debugger.go`:

```go
func NewGitHubWorkflowDebugger(apiKey string) *GitHubWorkflowDebugger {
    client := openai.NewClient(apiKey)

    model := os.Getenv("OPENAI_MODEL")
    if model == "" {
        model = openai.GPT4o  // Change this line
    }

    return &GitHubWorkflowDebugger{
        openaiClient: client,
        apiKey:       apiKey,
        model:        model,
    }
}
```

## Future Models

### GPT-5 (When Available)

When OpenAI releases GPT-5 models, you can use them immediately:

```bash
# GPT-5 mini (hypothetical)
export OPENAI_MODEL="gpt-5-mini"
./github-workflow-debugger <url>

# GPT-5 full (hypothetical)
export OPENAI_MODEL="gpt-5"
./github-workflow-debugger <url>
```

The agent is designed to work with any OpenAI model string, so it will automatically support new models as they're released.

## Model Selection Recommendations

### For Daily Use
```bash
OPENAI_MODEL="gpt-4o-mini"  # Default, best value
```

### For Important Failures
```bash
OPENAI_MODEL="gpt-4o"  # Better analysis
```

### For Cost Optimization
```bash
OPENAI_MODEL="gpt-4o-mini"  # Cheapest, still good quality
```

### For Maximum Quality (When Available)
```bash
OPENAI_MODEL="gpt-5"  # Future, when released
```

## Cost Comparison

| Model | Input Cost | Output Cost | Typical Analysis |
|-------|------------|-------------|------------------|
| gpt-4o-mini | $0.15/1M | $0.60/1M | $0.01-0.05 |
| gpt-4o | $2.50/1M | $10.00/1M | $0.15-0.50 |
| gpt-4-turbo | $10.00/1M | $30.00/1M | $0.50-1.50 |

*Costs as of November 2024, check https://openai.com/api/pricing/ for current rates*

## Model Capabilities

All models can:
- ✅ Analyze workflow logs
- ✅ Identify root causes
- ✅ Propose fixes
- ✅ Suggest code changes
- ✅ Rate confidence

Quality differences:
- **gpt-4o-mini**: Good insights, may miss edge cases
- **gpt-4o**: Excellent insights, catches subtle issues
- **gpt-4-turbo**: Very good, between mini and full

## Troubleshooting

### "Model not found" error
The model you specified doesn't exist. Check:
1. Model name spelling
2. OpenAI's current model list: https://platform.openai.com/docs/models
3. Your API key has access to that model

### Using gpt-5-mini before it exists
If you try to use `gpt-5-mini` before it's released:
```
Error: The model `gpt-5-mini` does not exist
```

Wait for OpenAI to release it, or use a current model.

## Checking Available Models

You can check which models are available using the OpenAI API:

```bash
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  | jq '.data[].id' | grep gpt
```
