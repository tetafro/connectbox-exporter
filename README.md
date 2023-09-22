# ConnectBox Exporter

[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/tetafro/connectbox-exporter/master/LICENSE)
[![Github CI](https://img.shields.io/github/actions/workflow/status/tetafro/connectbox-exporter/push.yml)](https://github.com/tetafro/connectbox-exporter/actions)
[![Go Report](https://goreportcard.com/badge/github.com/tetafro/connectbox-exporter)](https://goreportcard.com/report/github.com/tetafro/connectbox-exporter)
[![Codecov](https://codecov.io/gh/tetafro/connectbox-exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/tetafro/connectbox-exporter)

A Prometheus exporter for ConnectBox routers used by Ziggo internet provider
in the Netherlands.

Client with full API: [connectbox](https://github.com/tetafro/connectbox).

Mostly copied from [compal_CH7465LG_py](https://github.com/ties/compal_CH7465LG_py)
and [connectbox-prometheus](https://github.com/mbugert/connectbox-prometheus).

## Run

### Use docker

Create a config file `config.yml`
([example](https://github.com/tetafro/connectbox-exporter/blob/master/config.example.yml)).

```sh
docker run -d \
    --volume /host-dir/config.yml:/etc/prometheus/connectbox-exporter.yml \
    --publish 9119:9119 \
    --name connectbox-exporter \
    ghcr.io/tetafro/connectbox-exporter:latest
```

### Download binary

Download and unpack latest [release](https://github.com/tetafro/connectbox-exporter/releases).

Create a config file `config.yml`
([example](https://github.com/tetafro/connectbox-exporter/blob/master/config.example.yml)).

Run
```sh
./connectbox-exporter -config config.yml
```

### Build from sources

Clone the repository
```sh
git clone git@github.com:tetafro/connectbox-exporter.git
cd connectbox-exporter
```

Copy and populate config
```sh
cp config.example.yml config.yml
```

Build and run
```sh
make build run
```

## Get metrics

Get exporter internal metrics
```sh
curl 'http://localhost:9119/metrics'
```

Get ConnectBox metrics
```sh
curl 'http://localhost:9119/probe?target=192.168.178.1'
```

## Metrics

| Name                              | Type  | Description         |
| --------------------------------- | ----- | ------------------- |
| `connect_box_cm_docsis_mode`      | gauge | DocSis mode        |
| `connect_box_cm_hardware_version` | gauge | Hardware version   |
| `connect_box_cm_mac_addr`         | gauge | MAC address        |
| `connect_box_cm_network_access`   | gauge | Network access     |
| `connect_box_cm_serial_number`    | gauge | Serial number      |
| `connect_box_cm_system_uptime`    | gauge | System uptime      |
| `connect_box_lan_client`          | gauge | LAN client         |
| `connect_box_oper_state`          | gauge | Operational state  |
| `connect_box_temperature`         | gauge | Temperature        |
| `connect_box_tunner_temperature`  | gauge | Tunner temperature |
| `connect_box_wan_ipv4_addr`       | gauge | WAN IPv4 address   |
| `connect_box_wan_ipv6_addr`       | gauge | WAN IPv6 address   |
