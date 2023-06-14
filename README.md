# switchbot-meter-exporter

Prometheus export for SwitchBot Meter

## How to use

### Usage

```
Usage:
  main [OPTIONS]

Application Options:
  -p, --port= Port number to listen, default: 8080

Help Options:
  -h, --help  Show this help message
```

### Environment value

| key                 | description |
| ------------------- | ----------- |
| SWITCHBOT_OPENTOKEN | open token  |
| SWITCHBOT_SECRETKEY | secret key  |

### Prometheus config

```yml
scrape_configs:
  - job_name: "switchbot-meter"
    scrape_interval: 1m
    metrics_path: /probe
    static_configs:
      - targets:
          - <device_id>
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: <address>:<port>
```

## How to build

- `go mod download`
- `go build`

## License

Copyright (c) 2023 Rokoucha

Released under the MIT license, see LICENSE.
