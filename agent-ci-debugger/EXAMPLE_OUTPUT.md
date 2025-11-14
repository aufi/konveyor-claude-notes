# Example Output from GitHub Workflow Debugger

This is an example of what the AI agent would produce when analyzing the workflow failure at:
https://github.com/konveyor/ci/actions/runs/19353355807

---

## Example Analysis Report

```markdown
# GitHub Workflow Failure Analysis Report

**Workflow URL**: https://github.com/konveyor/ci/actions/runs/19353355807
**Repository**: konveyor/ci
**Run ID**: 19353355807
**Conclusion**: failure

---

## Root Cause

Test timeout in the e2e-api-integration-tests job. Specifically, the analysis task (Task ID 15)
for the Daytrader application exceeded the test timeout threshold (approximately 1 hour) without
completing successfully.

## Detailed Analysis

The workflow failed with the following sequence of events:

1. **Failed Component**: The e2e-api-integration-tests job in the main-nightly workflow
2. **Test Location**: `analysis_test.go:209` - this is where the timeout was detected
3. **Failure Mechanism**:
   - A test was running an analysis on the Daytrader application
   - The test created Task ID 15 and began polling for its completion
   - The task endpoint `/hub/tasks/15` was polled every 5 seconds, consistently returning HTTP 200
   - Despite the successful HTTP responses, the task never transitioned to a completed state
   - After approximately 1 hour of polling (from ~03:30 to ~04:32), the test framework timeout was triggered
   - The test dumped task information to `tmp_output/Daytrader/_task_15` for debugging

4. **Context from Logs**:
   - The Hub service was responding successfully (all GET requests returned 200)
   - The task appears to be "stuck" in a running state
   - Multiple analysis tasks were created during the test run (IDs 1-15)
   - Earlier tasks (language-discovery, tech-discovery) completed successfully
   - This suggests the issue is specific to the Daytrader analysis task, not a systemic problem

5. **Possible Root Causes**:
   - The Daytrader application analysis is taking longer than the timeout allows
   - The analyzer pod/container processing Task 15 may have crashed or hung
   - Resource constraints (CPU/memory) may be preventing the analysis from completing
   - A deadlock or infinite loop in the analyzer code when processing this specific application
   - External dependency issues (e.g., Maven repository access if using Java analysis)

## Proposed Fix

### Immediate Actions

1. **Investigate Task 15 Details**:
   ```bash
   # Download the workflow artifacts
   gh run download 19353355807 --repo konveyor/ci

   # Examine the task dump
   cat analysis-tests-output/Daytrader/_task_15/task.json
   cat analysis-tests-output/Daytrader/_task_15/logs.txt
   ```

2. **Check Analyzer Pod Status**:
   - Look for the analyzer pod that was processing Task 15
   - Check its logs for errors, warnings, or signs of hanging
   - Examine resource usage (CPU/memory) at the time of failure

3. **Reproduce Locally**:
   ```bash
   # Set up local environment
   make start-minikube
   make install-tackle

   # Run the Daytrader analysis with verbose logging
   # Monitor analyzer pod logs in real-time
   kubectl logs -f -n konveyor-tackle deployment/tackle-hub
   ```

### Short-term Fixes

1. **Increase Timeout**: If the analysis is legitimately taking longer:
   ```go
   // In analysis_test.go or test configuration
   const analysisTimeout = 2 * time.Hour  // Increase from 1 hour
   ```

2. **Add Resource Limits**: Ensure analyzer has sufficient resources:
   ```yaml
   # In Tackle CR
   spec:
     analyzer_container_memory: 4Gi  # Increase if needed
     analyzer_container_cpu: 2000m   # Increase if needed
   ```

3. **Add Better Logging**: Enhance task status reporting:
   - Log task state transitions
   - Add periodic "still running" indicators
   - Report progress percentage if possible

### Long-term Solutions

1. **Implement Analysis Progress Tracking**:
   - Add progress reporting to the analyzer
   - Show which phase of analysis is running
   - Provide estimated completion time

2. **Add Task Health Checks**:
   - Implement periodic health checks for running tasks
   - Detect and restart hung analyzer pods
   - Set maximum task runtime limits

3. **Optimize Daytrader Analysis**:
   - Profile the Daytrader analysis to identify bottlenecks
   - Optimize rule evaluation or code parsing
   - Consider caching or incremental analysis

4. **Improve Test Reliability**:
   - Add retry logic for flaky tests
   - Implement better timeout handling
   - Add task cancellation on timeout

## Files to Check

- `go-konveyor-tests/analysis_test.go` (line 209) - Where the timeout occurred
- `hub/addon/task.go` - Task management and status updates
- `hub/addon/analyzer.go` - Analyzer addon orchestration
- `analyzer-lsp/provider/java_external_provider.go` - Java analysis logic (if Daytrader is Java)
- `.github/workflows/global-ci-bundle.yml` - Workflow configuration
- `tackle2-operator/config/tackle-cr.yaml` - Resource limits and configuration

## Suggested Code Changes

### Change 1: Increase Test Timeout

**File**: `go-konveyor-tests/analysis_test.go`

```diff
 func TestDaytraderAnalysis(t *testing.T) {
-    timeout := 1 * time.Hour
+    timeout := 2 * time.Hour  // Increase timeout for large applications
+    t.Logf("Analysis timeout set to %v", timeout)

     ctx, cancel := context.WithTimeout(context.Background(), timeout)
     defer cancel()
```

### Change 2: Add Progress Logging

**File**: `go-konveyor-tests/analysis_test.go`

```diff
 func waitForTask(taskID uint) error {
     ticker := time.NewTicker(5 * time.Second)
+    lastLog := time.Now()

     for {
         select {
         case <-ticker.C:
             task, err := getTask(taskID)
             if err != nil {
                 return err
             }
+
+            // Log progress every minute
+            if time.Since(lastLog) > time.Minute {
+                t.Logf("Task %d still running, state: %s, started: %s",
+                    taskID, task.State, task.Started)
+                lastLog = time.Now()
+            }

             if task.State == "Succeeded" {
                 return nil
```

### Change 3: Add Resource Limits to Test CR

**File**: `.github/workflows/global-ci-bundle.yml`

```diff
   spec:
     image_pull_policy: IfNotPresent
-    analyzer_container_memory: 0
-    analyzer_container_cpu: 0
+    analyzer_container_memory: 6Gi  # Explicit memory limit
+    analyzer_container_cpu: 3000m   # Explicit CPU limit
     feature_auth_required: false
     disable_maven_search: true
```

**Confidence Level**: High

This is a well-understood timeout issue with clear logs showing the polling behavior.
The root cause is either insufficient resources or a legitimate long-running analysis
that exceeds the timeout. The proposed fixes address both possibilities.

---

*Generated at 2025-11-14T12:00:00Z*
```

## Key Features Demonstrated

1. **Structured Analysis**: Clear sections for root cause, analysis, and fixes
2. **Actionable Commands**: Specific bash/kubectl commands to investigate
3. **Code Changes**: Concrete diffs showing exactly what to modify
4. **Multiple Solutions**: Both quick fixes and long-term improvements
5. **Confidence Rating**: Indicates reliability of the analysis
6. **File References**: Direct pointers to files that need attention

## How to Use This Tool

```bash
# Set your API key
export ANTHROPIC_API_KEY="your-key-here"

# Run the debugger
./github-workflow-debugger https://github.com/konveyor/ci/actions/runs/19353355807

# Or use the example script
./example-usage.sh https://github.com/konveyor/ci/actions/runs/19353355807
```

The tool will generate a similar report saved to a timestamped markdown file that you can:
- Share with your team
- Reference in GitHub issues
- Use as a starting point for fixes
- Archive for future reference
