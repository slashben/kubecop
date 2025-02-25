package exporters

// here we will have the functionality to export the alerts to the alert manager
// Path: pkg/exporters/alert_manager.go

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/armosec/kubecop/pkg/engine/rule"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/models"
)

type AlertManagerExporter struct {
	Host     string
	NodeName string
	client   *client.AlertmanagerAPI
}

func InitAlertManagerExporter(alertmanagerURL string) *AlertManagerExporter {
	if alertmanagerURL == "" {
		alertmanagerURL = os.Getenv("ALERTMANAGER_URL")
		if alertmanagerURL == "" {
			return nil
		}
	}
	// Create a new Alertmanager client
	cfg := client.DefaultTransportConfig().WithHost(alertmanagerURL)
	amClient := client.NewHTTPClientWithConfig(nil, cfg)
	hostName, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("failed to get hostname: %v", err))
	}

	return &AlertManagerExporter{
		client:   amClient,
		Host:     hostName,
		NodeName: os.Getenv("NODE_NAME"),
	}
}

func (ame *AlertManagerExporter) SendAlert(failedRule rule.RuleFailure) {
	myAlert := models.PostableAlert{
		StartsAt: strfmt.DateTime(time.Now()),
		EndsAt:   strfmt.DateTime(time.Now().Add(time.Hour)),
		Annotations: map[string]string{
			"summary": fmt.Sprintf("Rule '%s' in '%s' namespace '%s' failed", failedRule.Name(), failedRule.Event().PodName, failedRule.Event().Namespace),
			"message": failedRule.Error(),
		},
		Alert: models.Alert{
			GeneratorURL: "http://github.com/armosec/kubecop",
			Labels: map[string]string{
				"alertname":      "KubeCopRuleViolated",
				"rule_name":      failedRule.Name(),
				"container_id":   failedRule.Event().ContainerID,
				"container_name": failedRule.Event().ContainerName,
				"namespace":      failedRule.Event().Namespace,
				"pod_name":       failedRule.Event().PodName,
				"severity":       PriorityToStatus(failedRule.Priority()),
				"host":           ame.Host,
				"node_name":      ame.NodeName,
			},
		},
	}

	// Send the alert
	params := alert.NewPostAlertsParams().WithContext(context.Background()).WithAlerts(models.PostableAlerts{&myAlert})
	isOK, err := ame.client.Alert.PostAlerts(params)
	if err != nil {
		fmt.Println("Error sending alert:", err)
		return
	}
	if isOK == nil {
		fmt.Println("Alert was not sent successfully")
		return
	}

	fmt.Printf("Alert sent successfully: %v\n", isOK)
}
