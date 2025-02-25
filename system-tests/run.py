import sys
import requests
import subprocess
import time
from promtopic import plotprom

alert_manager_url = "http://localhost:9093/"
prometheus_url = "http://localhost:9090/"



def filter_alerts_by_label(alerts, label_key, label_value):
    filtered_alerts = [alert for alert in alerts if label_key in alert['labels'] and alert['labels'][label_key] == label_value]
    return filtered_alerts

def get_active_alerts(alertmanager_url):
    endpoint = f"{alertmanager_url}/api/v2/alerts?active=true"
    try:
        response = requests.get(endpoint)
        response.raise_for_status()
        alerts = response.json()
        return alerts
    except requests.exceptions.HTTPError as errh:
        print(f"Http Error: {errh}")
    except requests.exceptions.ConnectionError as errc:
        print(f"Error Connecting: {errc}")
    except requests.exceptions.Timeout as errt:
        print(f"Timeout Error: {errt}")
    except requests.exceptions.RequestException as err:
        print(f"OOps: Something Else: {err}")

def basic_alert_test(namespace="kubecop-test"):
    print("Running basic alert test")

    # Create the namespace
    subprocess.check_call(["kubectl", "create", "namespace", namespace])

    try:
        # Install nginx profile in kubernetes by applying the nginx profile yaml
        subprocess.check_call(["kubectl", "-n", namespace , "apply", "-f", "dev/nginx/nginx-app-profile.yaml"])

        # Install nginx in kubernetes by applying the nginx deployment yaml
        subprocess.check_call(["kubectl", "-n", namespace , "apply", "-f", "dev/nginx/nginx-deployment.yaml"])

        # Wait for nginx to be ready
        subprocess.check_call(["kubectl", "-n", namespace , "wait", "--for=condition=ready", "pod", "-l", "app=nginx", "--timeout=120s"])

        # Get the pod name of the nginx pod
        nginx_pod_name = subprocess.check_output(["kubectl", "-n", namespace , "get", "pod", "-l", "app=nginx", "-o", "jsonpath='{.items[0].metadata.name}'"]).decode("utf-8").strip("'")

        # Exec into the nginx pod and create a file in the /tmp directory
        subprocess.check_call(["kubectl", "-n", namespace , "exec", nginx_pod_name, "--", "touch", "/tmp/nginx-test"])

        print("Starting load on nginx pod")

        # Create load on the nginx pod
        subprocess.check_call(["kubectl", "-n", namespace , "apply", "-f", "system-tests/locust-deployment.yaml"])

        # Wait for the locust pod to be ready
        subprocess.check_call(["kubectl", "-n", namespace , "wait", "--for=condition=ready", "pod", "-l", "app=http-loader", "--timeout=120s"])

        print("Waiting 300 seconds to create load")

        # Wait for the alert to be fired
        time.sleep(300)
    except:
        # Delete the namespace
        subprocess.check_call(["kubectl", "delete", "namespace", namespace])
        return 1

    subprocess.check_call(["kubectl", "delete", "namespace", namespace])

    # Get the active alerts
    alerts = get_active_alerts(alert_manager_url)
    if not alerts:
        print("Could not get alerts")
        return 1

    alerts = filter_alerts_by_label(alerts, "alertname", "KubeCopRuleViolated")
    alerts = filter_alerts_by_label(alerts, "rule_name", "Exec Whitelisted")

    if len(alerts) == 0:
        print("No alerts found")
        return 1
    else:
        print("Found alerts %s" % alerts)
        return 0

def rule_binding_apply_test(namespace="kubecop-test"):
    print("Running rule binding apply test")

    try:        
        subprocess.check_call(["kubectl", "apply", "-f", "system-tests/rule_binding_crds_files/all-valid.yaml"])
        subprocess.check_call(["kubectl", "delete", "-f", "system-tests/rule_binding_crds_files/all-valid.yaml"])
        # invalid fields
        proc_stat = subprocess.run(["kubectl", "apply", "-f", "system-tests/rule_binding_crds_files/invalid-name.yaml"], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if proc_stat.returncode == 0:
            print("Invalid name test failed")
            return 1
        
        proc_stat = subprocess.run(["kubectl", "apply", "-f", "system-tests/rule_binding_crds_files/invalid-id.yaml"], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if proc_stat.returncode == 0:
            print("Invalid id test failed")
            return 1
        
        proc_stat = subprocess.run(["kubectl", "apply", "-f", "system-tests/rule_binding_crds_files/invalid-tag.yaml"], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if proc_stat.returncode == 0:
            print("Invalid tag test failed")
            return 1
        
        # duplicate fields
        proc_stat = subprocess.run(["kubectl", "apply", "-f", "system-tests/rule_binding_crds_files/dup-fields-name-tag.yaml"], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if proc_stat.returncode == 0:
            print("Duplicate fields name-tag test failed")
            return 1
        
        proc_stat = subprocess.run(["kubectl", "apply", "-f", "system-tests/rule_binding_crds_files/dup-fields-name-id.yaml"], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if proc_stat.returncode == 0:
            print("Duplicate fields name-id test failed")
            return 1
        
        proc_stat = subprocess.run(["kubectl", "apply", "-f", "system-tests/rule_binding_crds_files/dup-fields-id-tag.yaml"], check=False, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if proc_stat.returncode == 0:
            print("Duplicate fields id-tag test failed")
            return 1

    except:
        return 1


test_cases = [
    (basic_alert_test, "Basic alert test"),
    (rule_binding_apply_test, "Rule binding apply test")
]

def main():
    global alert_manager_url
    global prometheus_url
    if len(sys.argv) > 1:
        alert_manager_url = sys.argv[1]
    if len(sys.argv) > 2:
        prometheus_url = sys.argv[2]
    print("Running tests")
    for test_case, test_case_name in test_cases:
        print("Running test %s" % test_case_name)
        # Save start time in epoch
        time_start = time.time()
        result = test_case()
        # Save end time in epoch
        time_end = time.time()
        if result == 0:
            print("Test passed")
        else:
            print("Test failed")
            return result
        # Give two minutes for prometheus to scrape the data
        print("Waiting 60 seconds for prometheus to scrape the data")
        time.sleep(60)
        result = plotprom(test_case_name, time_start, time_end)
        if result == 0:
            print("Ploting succeeded")
        else:
            print("Ploting failed")
    sys.exit(result)


if __name__ == "__main__":
    main()