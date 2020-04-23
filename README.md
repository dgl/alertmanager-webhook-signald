# Signal notifier for Prometheus alertmanager

This implements an alertmanager webhook that can connect to
[signald](https://github.com/thefinn93/signald).

## Setup

Follow https://github.com/thefinn93/signald#quick-start and get a number
registered in signald. Set this number as the 'sender' in the config below.

## Configuration

```yaml
defaults:
  # Phone number of sender, must be registered in this signald per Setup.
  sender: +1xxx
  template: '{{ template "signal.message" . }}'
  # Subscribe to responses from signald. May help to keep the connection alive.
  subscribe: true

templates:
  - "alerts.tmpl"

receivers:
  - name: something
    to:
      - group:xxxx
      - tel:+44...
    # Optional: the sender, template, etc. fields as in defaults above.
```

See [example.yaml](example.yaml) for a more complete configuration example.

### Alertmanager configuration

```yaml
receivers:
  - name: something
    webhook_configs:
      - url: http://localhost:9716/alert
```

The receiver name defined in alertmanager configuration will be sent to the
receiver with the matching name in the receivers section of the configuration
file (i.e. "something" in this example must be the same string in both
alertmanager configuration and this webhook's configuration).

### Prometheus configuration

Use Prometheus to check the health of the webhook itself.

Configure Prometheus to scrape it:
```yaml
scrape_configs:
  - job_name: alertmanager-signald-webhook
    static_configs:
      - targets: ['localhost:9716']
```

Configure some rules like the rules in [example-rules.yaml](example-rules.yaml)
to alert you ideally via another alert receiver!
