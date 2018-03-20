## Custom Script Result exporter (csr_exporter) for Prometheus
Prometheus exporter for custom user defined scripts. Allow user to list multiple command which will be executed via `sh` and result will be parsed as float and then exported to prometheus via HTTP end point.

### How to run
```bash
./csr_exporter -config ./csr_exporter.toml
```
To enable debug output - use `-verbose`
```bash
./csr_exporter -config ./csr_exporter.toml -verbose
```

### Configuration file example
```toml
update_period_sec = 5
port = 8080

[[metrics]]
name = "system_cpu_temp"
help = "CPU temperature"
cmd = "cat /sys/class/thermal/thermal_zone0/temp | cut -c-2"

[[metrics]]
name = "system_cpu_temp2"
help = "CPU temperature"
cmd = "cat /tmp/temp_test"
```

### How to build
1. Install dependencies. We use `dep` (https://golang.github.io/dep/docs/installation.html) as dependencies manager for go.
```bash
~/go/bin/dep ensure
```

2. Build
```bash
go build
```
