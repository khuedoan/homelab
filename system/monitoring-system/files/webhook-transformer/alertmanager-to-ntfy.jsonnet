local get_tags(status, severity) =
  // https://docs.ntfy.sh/emojis
  if status == "resolved" then
    ["tada"]
  else
    std.get({
      critical: ["rotating_light"],
      warning: ["warning"],
      info: ["newspaper"],
    }, severity, ["question"]);

local get_priority(status, severity) =
  // https://docs.ntfy.sh/publish/#message-priority
  if status == "resolved" then
    2
  else
    std.get({
      critical: 5,
      warning: 3,
      info: 1,
    }, severity, 3);

local get_actions(status, annotations) =
  // https://docs.ntfy.sh/publish/#action-buttons
  if status == "resolved" || !("runbook_url" in annotations) then
    []
  else
    [
      {
        action: "view",
        label: "Open runbook",
        url: annotations.runbook_url,
      },
    ];

// TODO support multiple alerts
{
  topic: env.NTFY_TOPIC,
  title: "[" + std.asciiUpper(body.status) + "] " + body.alerts[0].labels.alertname,
  message: body.alerts[0].annotations.description,
  tags: get_tags(body.status, body.alerts[0].labels.severity),
  priority: get_priority(body.status, body.alerts[0].labels.severity),
  actions: get_actions(body.status, body.alerts[0].annotations),
}
