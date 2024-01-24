{
  "topic": env.NTFY_TOPIC,
  "title": body.alerts[0].labels.alertname, // TODO support multiple alerts
  "message": body.alerts[0].annotations.description,
  "tags": [],
  "priority": 3,
  "actions": []
}
