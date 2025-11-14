#!/bin/bash

# Example usage of the GitHub Workflow Debugger
#
# This script demonstrates how to use the workflow debugger to analyze
# a failed GitHub Actions workflow run.

# Ensure OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "Error: OPENAI_API_KEY environment variable is not set"
    echo "Please set it with: export OPENAI_API_KEY='your-api-key'"
    exit 1
fi

# Ensure gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if gh is authenticated
if ! gh auth status &> /dev/null; then
    echo "Error: GitHub CLI is not authenticated"
    echo "Run: gh auth login"
    exit 1
fi

# Build the debugger if it doesn't exist
if [ ! -f ./github-workflow-debugger ]; then
    echo "Building github-workflow-debugger..."
    go build -o github-workflow-debugger github-workflow-debugger.go
    if [ $? -ne 0 ]; then
        echo "Error: Failed to build github-workflow-debugger"
        exit 1
    fi
fi

# Example workflow URLs to debug
# Uncomment one or provide your own

# Example 1: Konveyor CI nightly test failure
WORKFLOW_URL="https://github.com/konveyor/ci/actions/runs/19353355807"

# Example 2: Use command line argument if provided
if [ $# -gt 0 ]; then
    WORKFLOW_URL="$1"
fi

echo "============================================"
echo "GitHub Workflow Debugger"
echo "============================================"
echo "Analyzing: $WORKFLOW_URL"
echo "============================================"
echo ""

# Run the debugger
./github-workflow-debugger "$WORKFLOW_URL"

# Check if the report was generated
if [ $? -eq 0 ]; then
    echo ""
    echo "============================================"
    echo "Analysis complete!"
    echo "Report saved to: workflow-debug-*.md"
    echo "============================================"

    # Optionally open the report
    LATEST_REPORT=$(ls -t workflow-debug-*.md 2>/dev/null | head -1)
    if [ -n "$LATEST_REPORT" ]; then
        echo ""
        read -p "Would you like to view the report? (y/n) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if command -v bat &> /dev/null; then
                bat "$LATEST_REPORT"
            elif command -v less &> /dev/null; then
                less "$LATEST_REPORT"
            else
                cat "$LATEST_REPORT"
            fi
        fi
    fi
else
    echo ""
    echo "============================================"
    echo "Error: Analysis failed"
    echo "============================================"
    exit 1
fi
