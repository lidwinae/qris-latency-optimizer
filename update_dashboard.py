import json

with open('grafana/dashboards/golang-metrics.json', 'r') as f:
    data = json.load(f)

# Find max gridPos y
max_y = 0
for panel in data['panels']:
    y = panel.get('gridPos', {}).get('y', 0)
    h = panel.get('gridPos', {}).get('h', 0)
    if y + h > max_y:
        max_y = y + h

new_panels = [
    {
        "type": "row",
        "title": "CPU & Resource Monitoring",
        "gridPos": {"x": 0, "y": max_y, "w": 24, "h": 1},
        "id": 900
    },
    {
        "title": "Go Backend CPU Usage",
        "type": "timeseries",
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "gridPos": {"x": 0, "y": max_y + 1, "w": 8, "h": 8},
        "id": 901,
        "targets": [{"expr": "rate(process_cpu_seconds_total[1m])", "legendFormat": "Backend CPU (Cores)"}]
    },
    {
        "title": "Docker Containers CPU Usage",
        "type": "timeseries",
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "gridPos": {"x": 8, "y": max_y + 1, "w": 8, "h": 8},
        "id": 902,
        "targets": [{"expr": "sum(rate(container_cpu_usage_seconds_total{image!=\"\", container_label_com_docker_compose_project=\"qris-latency-optimizer\"}[1m])) by (name)", "legendFormat": "{{name}}"}]
    },
    {
        "title": "Docker Containers RAM Usage",
        "type": "timeseries",
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "gridPos": {"x": 16, "y": max_y + 1, "w": 8, "h": 8},
        "id": 903,
        "targets": [{"expr": "sum(container_memory_usage_bytes{image!=\"\", container_label_com_docker_compose_project=\"qris-latency-optimizer\"}) by (name)", "legendFormat": "{{name}}"}]
    }
]

data['panels'].extend(new_panels)

with open('grafana/dashboards/golang-metrics.json', 'w') as f:
    json.dump(data, f, indent=2)

