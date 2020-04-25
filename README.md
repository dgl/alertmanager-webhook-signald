# Signal notifier for Prometheus alertmanager

This implements an alertmanager webhook that can connect to
[signald](https://github.com/thefinn93/signald).

## Building

```shell
go get github.com/dgl/alertmanager-webhook-signald
```

Will give you a *alertmanager-webhook-signald* binary to run, in your Go bin
directory (`$(go env GOPATH)/bin/alertmanager-webhook-signald`).

## Setup

Follow https://github.com/thefinn93/signald#quick-start and get a number
registered in signald. Set this number as the 'sender' in the config below.

## Configuration

Save something like the following as *config.yaml*:

```yaml
defaults:
  # Phone number of sender, must be registered in this signald per Setup.
  sender: +1xxx
  template: '{{ template "signal.message" . }}'
  # Subscribe to responses from signald. May help to keep the connection alive.
  subscribe: true

templates:
  # Copy this file to the same place as the configuration file.
  - "alerts.tmpl"

receivers:
  - name: something
    to:
      - group:xxxx
      - tel:+44...
    # Optional: the sender, template, etc. fields as in defaults above.
```

See [example.yaml](example.yaml) for a more complete configuration example.

You'll also need a template file for the alert message text, just putting
[alerts.tmpl](alerts.tmpl) in the same directory as the configuration file will
work for most cases.

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

## Running

```shell
alertmanager-webhook-signald -config config.yaml
```

## Monitoring

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
to alert you -- ideally via another alert receiver!
