# Signal notifier for Prometheus alertmanager

This implements an alertmanager webhook that can connect to
[signald](https://github.com/thefinn93/signald).

Pretty alpha right now (in particular signald is beta).

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

### Alertmanager configuration

```yaml
receivers:
  - name: something
    webhook_configs:
      - url: http://localhost:9245/alert
```

The receiver name defined in alertmanager configuration will be sent to the
receiver with the matching name in the receivers section of the configuration
file (i.e. "something" in this example must be the same string in both
alertmanager configuration and this webhook's configuration).
