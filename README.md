# K8sLogmonitor

Kubernetes Logs Monitoring Using Agent Development Kit (ADK) for Go with AI-powered failure detection and recommendations.

## Overview

K8sLogmonitor is an intelligent Kubernetes monitoring solution that:
- **Monitors pod logs** across all namespaces
- **Detects failures** using pattern matching
- **Generates AI recommendations** via OpenAI API
- **Provides web UI** for interactive monitoring
- **Follows ADK patterns** for extensible agent architecture

## Features

### 🔍 Comprehensive Failure Detection
- Image pull failures (ImagePullBackOff)
- Container crashes (CrashLoopBackOff)
- Resource issues (OOMKilled, CPU throttling)
- Health check failures (readiness/liveness probes)
- Network connectivity issues
- Storage and volume mount problems
- Security and permission errors

### 🤖 AI-Powered Recommendations
- **OpenAI Integration**: Real-time troubleshooting advice
- **GitHub Issues Lookup**: Automatic search for similar issues in repository
- **Fallback Recommendations**: Built-in solutions for common issues
- **Contextual Analysis**: Includes pod status, events, and resource info

### 🌐 Web Interface
- **Individual Pod Monitoring**: Target specific pods
- **All-Namespace Scanning**: Monitor entire cluster
- **Real-time Results**: Instant failure detection and recommendations
- **Clean UI**: Failed pods with actionable advice

### 🏗️ ADK Architecture
- **Agent-based Design**: Modular and extensible
- **Tool Registry**: Pluggable tool system
- **Standardized Interfaces**: Consistent tool execution

## Installation

### Prerequisites
- Go 1.25+
- Kubernetes cluster access
- OpenAI API key (optional, has fallbacks)

### Setup

1. **Clone Repository**
   ```bash
   git clone https://github.com/vasudevchavan/K8sLogmonitor
   cd K8sLogmonitor
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure Kubernetes Access**
   ```bash
   # Ensure kubeconfig is available
   export KUBECONFIG=~/.kube/config
   ```

4. **Set API Keys** (Optional)
   ```bash
   export LLM_API_KEY="your-openai-api-key"
   export GITHUB_TOKEN="your-github-token"  # For higher rate limits
   ```

## Usage

### Command Line Monitoring

**Step 1: Set Environment Variables**
```bash
export LLM_API_KEY="your-openai-api-key"
export GITHUB_TOKEN="your-github-token"  # Optional
export KUBECONFIG=~/.kube/config
```

**Step 2: Run Continuous Monitoring**
```bash
go run main.go
```

**Step 3: Web Interface (Alternative)**
```bash
go run cmd/web/main.go
```
Access: http://localhost:8080

### Web UI Features

#### Individual Pod Monitoring
1. Enter namespace, pod name, and container name
2. Click "🔍 Monitor Pod"
3. View failures and AI recommendations

#### All-Namespace Monitoring
1. Click "🌐 Monitor All Namespaces"
2. System scans all pods across all namespaces
3. Displays only failed pods with recommendations

## Architecture

### ADK Components

```
├── adk/
│   └── agent.go          # ADK interfaces and base agent
├── agents/
│   ├── log_monitor_agent.go    # Main monitoring agent
│   ├── pod_log_agent.go        # Pod log fetching
│   ├── failure_detection_agent.go  # Pattern matching
│   └── recommendation_agent.go     # AI recommendations
├── tools/
│   ├── k8s_tool.go        # Kubernetes operations
│   ├── k8s_context_tool.go    # Pod context gathering
│   ├── llm_tool.go        # OpenAI integration
│   └── getpodlogs.go      # Log retrieval
├── web/
│   ├── server.go          # Web server
│   └── api.go            # REST API endpoints
└── config/
    └── thresholds.go      # Configuration
```

### Tool Registry

- **k8s_logs**: Fetches pod logs
- **k8s_context**: Gathers pod metadata, events, resources
- **failure_detection**: Pattern-based failure detection
- **github_issues**: Searches repository for similar issues
- **llm_recommendation**: AI-powered recommendations with GitHub context

## API Endpoints

### POST /api/monitor
Monitor specific pod
```json
{
  "namespace": "default",
  "pod_name": "my-pod",
  "container_name": "my-container"
}
```

### GET /api/monitor-all
Scan all namespaces for failures
```json
[
  {
    "namespace": "default",
    "pod_name": "failed-pod",
    "container_name": "app",
    "failures": "Image pull error",
    "recommendation": "Check image name and registry access"
  }
]
```

## Configuration

### Environment Variables
- `LLM_API_KEY`: OpenAI API key for AI recommendations
- `GITHUB_TOKEN`: GitHub personal access token (optional, for higher rate limits)
- `KUBECONFIG`: Path to Kubernetes config file

### Thresholds
```go
type Thresholds struct {
    LogTailLines      int64  // Number of log lines to fetch
    MonitorIntervalMs int    // Monitoring interval
    MaxFailuresCount  int    // Max failures to process
}
```

## Example Output

### Command Line Output

**Startup:**
```
I1114 00:52:30.915413   40270 k8s_client.go:28] kubeconfig file was loaded from ~/home/.kube/config
```

**Monitoring Results:**
```
2025/11/14 00:53:30 DEBUG: Detected 0 failures for test/test: []
2025/11/14 00:53:30 DEBUG: Detected 3 failures for testin1g/testin1g: [error: pull image waiting to start]
2025/11/14 00:53:30 DEBUG: Searching GitHub for: testin1g image
2025/11/14 00:53:31 DEBUG: Found 1 GitHub issues
2025/11/14 00:53:31 DEBUG: GitHub issues formatted: Related GitHub Issues:
1. Issue #1: Pod testin2g failure wrong image (open)
   Description: Image need to be updated for **testin2g** pod
   URL: https://github.com/vasudevchavan/K8sLogmonitor/issues/1
2025/11/14 00:53:31 DEBUG: Calling LLM with enhanced context including GitHub issues
2025/11/14 00:53:32 DEBUG: LLM recommendation: Image Pull Error - Check: 1) Image name/tag correctness 2) Registry accessibility 3) Image pull secrets 4) Network connectivity
```

**Final Recommendation:**
```
Pod: testin1g/testin1g
Failures: [error: pull image waiting to start]
Recommendation: Image Pull Error - Check: 1) Image name/tag correctness 2) Registry accessibility 3) Image pull secrets 4) Network connectivity
```

**System Behavior:**
- ✅ **Healthy Pods**: Silent monitoring (test/test, testing/testing)
- ❌ **Failed Pods**: Detailed analysis with GitHub issue lookup
- 🔍 **GitHub Integration**: Automatic search finds Issue #1 for testin2g pod
- 🤖 **AI Recommendations**: Contextual troubleshooting steps
- ⏱️ **Continuous**: Monitors every minute, Ctrl+C to stop

### Web Interface
- **Failed Pods**: Red-highlighted containers with issues
- **Failure Details**: Specific error patterns detected
- **AI Recommendations**: Actionable troubleshooting steps with GitHub issue references
- **GitHub Integration**: Automatic lookup of related repository issues
- **All Clear**: Green message when no failures found

## Troubleshooting

### Common Issues

**Port 8080 in use:**
```bash
lsof -ti:8080 | xargs kill -9
```

**OpenAI Rate Limits:**
- System provides fallback recommendations with GitHub issue links
- Consider upgrading OpenAI plan

**GitHub API Rate Limits:**
- Anonymous: 60 requests/hour
- With token: 5000 requests/hour
- Set GITHUB_TOKEN for higher limits

**Kubernetes Access:**
```bash
kubectl cluster-info
kubectl auth can-i get pods --all-namespaces
```

**Running the Application:**
```bash
# 1. Ensure Kubernetes access
kubectl get pods --all-namespaces

# 2. Set API keys (optional but recommended)
export LLM_API_KEY="sk-..."
export GITHUB_TOKEN="ghp_..."

# 3. Run the monitor
go run main.go

# 4. Stop with Ctrl+C
```

## GitHub Issues Integration

### How It Works
1. **Automatic Search**: When failures are detected, system searches your repository for similar issues
2. **Smart Matching**: Searches both issue titles and body content using pod names and error types
3. **Contextual Recommendations**: LLM receives GitHub issue context for better troubleshooting advice
4. **Fallback Integration**: Even without LLM, recommendations include relevant GitHub issue links

### Creating Helpful Issues
To maximize the GitHub integration benefits:

```markdown
# Issue Title Format
Pod [pod-name] failure [error-type]

# Example
Pod testin2g failure wrong image

# Description
Include:
- Pod name and namespace
- Error symptoms
- Resolution steps
- Related configuration
```

### Repository Configuration
By default, searches `vasudevchavan/K8sLogmonitor`. To use your own repository:

1. Update `agents/log_monitor_agent.go` line with your repo:
   ```go
   "repo": "your-org/your-repo",
   ```

2. Update `tools/github_tool.go` default repo:
   ```go
   repo = "your-org/your-repo"
   ```

## Contributing

1. Fork the repository
2. Create feature branch
3. Follow ADK patterns for new tools/agents
4. Add tests for new functionality
5. Create GitHub issues for bugs/features to test integration
6. Submit pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- GitHub Issues: [Create Issue](https://github.com/vasudevchavan/K8sLogmonitor/issues)
- Documentation: See code comments and examples
