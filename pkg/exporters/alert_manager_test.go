package exporters

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/armosec/kubecop/pkg/engine/rule"
	"github.com/kubescape/kapprofiler/pkg/tracing"
	"github.com/stretchr/testify/assert"
)

func TestSendAlert(t *testing.T) {
	// Set up a mock Alertmanager server
	recievedData := make(chan []byte, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		bodyData, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}
		recievedData <- bodyData
	}))
	defer server.Close()
	// os.Setenv("ALERTMANAGER_URL", "localhost:9093")

	// Create a new Alertmanager exporter
	exporter := InitAlertManagerExporter(strings.Replace(server.URL, "http://", "", 1))
	if exporter == nil {
		t.Fatalf("Failed to create new Alertmanager exporter")
	}
	// Call SendAlert

	exporter.SendAlert(&rule.R0001ExecWhitelistedFailure{
		RuleName: "testrule",
		Err:      "Application profile is missing",
		FailureEvent: &tracing.ExecveEvent{GeneralEvent: tracing.GeneralEvent{
			ContainerName: "testcontainer", ContainerID: "testcontainerid", Namespace: "testnamespace", PodName: "testpodname"}},
	})
	bytesData := <-recievedData

	// Assert the request body is correct
	alerts := []map[string]interface{}{}
	if err := json.Unmarshal(bytesData, &alerts); err != nil {
		t.Fatalf("Failed to unmarshal request body: %v", err)
	}
	assert.Equal(t, 1, len(alerts))
	alert := alerts[0]
	alertLabels := alert["labels"].(map[string]interface{})
	assert.Equal(t, "KubeCopRuleViolated", alertLabels["alertname"])
	assert.Equal(t, "testrule", alertLabels["rule_name"])
	assert.Equal(t, "testcontainerid", alertLabels["container_id"])
	assert.Equal(t, "testcontainer", alertLabels["container_name"])
	assert.Equal(t, "testnamespace", alertLabels["namespace"])
	assert.Equal(t, "testpodname", alertLabels["pod_name"])
	assert.Equal(t, "", alertLabels["node_name"])
	assert.Equal(t, "none", alertLabels["severity"])
	assert.Equal(t, "Rule 'testrule' in 'testpodname' namespace 'testnamespace' failed", alert["annotations"].(map[string]interface{})["summary"])
	assert.Equal(t, "Application profile is missing", alert["annotations"].(map[string]interface{})["message"])
	assert.Equal(t, "http://github.com/armosec/kubecop", alert["generatorURL"])

}
