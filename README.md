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
  template: "{{ template ...}}"

templates:
  - "alerts.tmpl"

receivers:
  - name: something
    to:
      - group:xxxx
      - tel:+44...
    # Optional: the sender, template fields as in defaults above.
```
