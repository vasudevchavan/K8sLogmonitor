package web

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/vasudevchavan/K8sLogmonitor/adk"
	"github.com/vasudevchavan/K8sLogmonitor/agents"
	"github.com/vasudevchavan/K8sLogmonitor/tools"
	"k8s.io/client-go/kubernetes"
)

type Server struct {
	agent     *agents.LogMonitorAgent
	k8sClient *kubernetes.Clientset
}

type MonitorRequest struct {
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
}

type MonitorResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
	Error   string `json:"error,omitempty"`
}

func NewServer() (*Server, error) {
	k8sClient, err := tools.NewK8sClient()
	if err != nil {
		return nil, err
	}

	registry := adk.NewToolRegistry()
	registry.RegisterTool("k8s_logs", tools.NewK8sTool(k8sClient))
	registry.RegisterTool("k8s_context", tools.NewK8sContextTool(k8sClient))
	registry.RegisterTool("failure_detection", tools.NewFailureDetectionTool())
	registry.RegisterTool("github_issues", tools.NewGitHubTool(os.Getenv("GITHUB_TOKEN")))
	registry.RegisterTool("llm_recommendation", tools.NewLLMTool(os.Getenv("LLM_API_KEY")))

	agent := agents.NewLogMonitorAgent(registry)

	return &Server{
		agent:     agent,
		k8sClient: k8sClient,
	}, nil
}

func (s *Server) monitorHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	input := strings.Join([]string{req.Namespace, req.PodName, req.ContainerName}, "|")
	result, err := s.agent.Execute(context.Background(), input)

	resp := MonitorResponse{
		Success: err == nil,
		Result:  result,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>K8s Log Monitor</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 800px; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        input, button { padding: 12px; margin: 8px; border: 1px solid #ddd; border-radius: 5px; font-size: 14px; }
        input { width: 200px; }
        button { background: #007cba; color: white; cursor: pointer; border: none; }
        button:hover { background: #005a87; }
        .result { background: #f8f9fa; padding: 20px; margin: 20px 0; border-radius: 8px; border-left: 4px solid #007cba; }
        .error { background: #ffebee; border-left-color: #f44336; }
        .success { background: #e8f5e8; border-left-color: #4caf50; }
        pre { white-space: pre-wrap; word-wrap: break-word; }
        h1 { color: #333; margin-bottom: 30px; }
        .form-group { margin: 15px 0; }
        label { display: inline-block; width: 120px; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 K8s Log Monitor</h1>
        <form id="monitorForm">
            <div class="form-group">
                <label>Namespace:</label>
                <input type="text" id="namespace" placeholder="default" value="default">
            </div>
            <div class="form-group">
                <label>Pod Name:</label>
                <input type="text" id="podName" placeholder="Enter pod name" required>
            </div>
            <div class="form-group">
                <label>Container:</label>
                <input type="text" id="containerName" placeholder="Enter container name" required>
            </div>
            <button type="submit">🔍 Monitor Pod</button>
            <button type="button" onclick="monitorAll()">🌐 Monitor All Namespaces</button>
        </form>
        <div id="result"></div>
    </div>

    <script>
        document.getElementById('monitorForm').onsubmit = async function(e) {
            e.preventDefault();
            
            const resultDiv = document.getElementById('result');
            resultDiv.innerHTML = '<div class="result">⏳ Analyzing pod...</div>';
            
            const data = {
                namespace: document.getElementById('namespace').value || 'default',
                pod_name: document.getElementById('podName').value,
                container_name: document.getElementById('containerName').value
            };

            try {
                const response = await fetch('/api/monitor', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });

                const result = await response.json();
                
                if (result.success) {
                    resultDiv.className = 'result success';
                    resultDiv.innerHTML = '<h3>✅ Analysis Complete:</h3><pre>' + result.result + '</pre>';
                } else {
                    resultDiv.className = 'result error';
                    resultDiv.innerHTML = '<h3>❌ Error:</h3><p>' + result.error + '</p>';
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error"><h3>❌ Request Failed:</h3><p>' + error.message + '</p></div>';
            }
        };

        async function monitorAll() {
            const resultDiv = document.getElementById('result');
            resultDiv.innerHTML = '<div class="result">⏳ Scanning all namespaces...</div>';
            
            try {
                const response = await fetch('/api/monitor-all');
                const failures = await response.json();
                
                if (failures.length === 0) {
                    resultDiv.innerHTML = '<div class="result success"><h3>✅ All Clear!</h3><p>No pod failures detected across all namespaces.</p></div>';
                } else {
                    let html = '<div class="result"><h3>🚨 Failed Pods (' + failures.length + '):</h3>';
                    failures.forEach(failure => {
                        html += '<div style="margin: 20px 0; padding: 20px; background: #ffebee; border-left: 4px solid #f44336; border-radius: 8px;">';
                        html += '<h4 style="margin: 0 0 10px 0; color: #d32f2f;">🚫 ' + failure.namespace + '/' + failure.pod_name + '/' + failure.container_name + '</h4>';
                        html += '<div style="background: #fff; padding: 10px; border-radius: 4px; margin: 10px 0;"><strong>Failures:</strong><br>' + failure.failures.replace(/Failures: /, '') + '</div>';
                        html += '<div style="background: #e3f2fd; padding: 10px; border-radius: 4px; border-left: 3px solid #2196f3;"><strong>💡 Recommendation:</strong><br>' + failure.recommendation + '</div>';
                        html += '</div>';
                    });
                    html += '</div>';
                    resultDiv.innerHTML = html;
                }
            } catch (error) {
                resultDiv.innerHTML = '<div class="result error"><h3>❌ Scan Failed:</h3><p>' + error.message + '</p></div>';
            }
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (s *Server) Start(port string) error {
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/api/monitor", s.monitorHandler)
	http.HandleFunc("/api/monitor-all", s.monitorAllHandler)
	return http.ListenAndServe(":"+port, nil)
}