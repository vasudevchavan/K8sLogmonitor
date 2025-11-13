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

4. **Set OpenAI API Key** (Optional)
   ```bash
   export LLM_API_KEY="your-openai-api-key"
   ```

## Usage

### Command Line Monitoring

**Continuous Monitoring:**
```bash
go run main.go
```

**Web Interface:**
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
- **llm_recommendation**: AI-powered recommendations

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

### Command Line
```
DEBUG: Detected 3 failures for testin1g/testin1g: [error: pull image waiting to start]
DEBUG: Calling LLM with enhanced context
Pod: testin1g/testin1g
Failures: [error: pull image waiting to start]
Recommendation: Image Pull Error - Check: 1) Image name/tag correctness 2) Registry accessibility 3) Image pull secrets 4) Network connectivity
```

### Web Interface
- **Failed Pods**: Red-highlighted containers with issues
- **Failure Details**: Specific error patterns detected
- **AI Recommendations**: Actionable troubleshooting steps
- **All Clear**: Green message when no failures found

## Troubleshooting

### Common Issues

**Port 8080 in use:**
```bash
lsof -ti:8080 | xargs kill -9
```

**OpenAI Rate Limits:**
- System provides fallback recommendations
- Consider upgrading OpenAI plan

**Kubernetes Access:**
```bash
kubectl cluster-info
kubectl auth can-i get pods --all-namespaces
```

## Contributing

1. Fork the repository
2. Create feature branch
3. Follow ADK patterns for new tools/agents
4. Add tests for new functionality
5. Submit pull request

## License

MIT License - see LICENSE file for details.

## Support

For issues and questions:
- GitHub Issues: [Create Issue](https://github.com/vasudevchavan/K8sLogmonitor/issues)
- Documentation: See code comments and examples
