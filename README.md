# ConnectBox Exporter

[![License](http://img.shields.io/badge/license-MIT-green.svg?style=flat)](https://raw.githubusercontent.com/tetafro/connectbox-exporter/master/LICENSE)
[![Github CI](https://img.shields.io/github/actions/workflow/status/tetafro/connectbox-exporter/push.yml)](https://github.com/tetafro/connectbox-exporter/actions)
[![Go Report](https://goreportcard.com/badge/github.com/tetafro/connectbox-exporter)](https://goreportcard.com/report/github.com/tetafro/connectbox-exporter)
[![Codecov](https://codecov.io/gh/tetafro/connectbox-exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/tetafro/connectbox-exporter)

**WORK IN PROGRESS**

A Prometheus exporter for ConnectBox routers used by Ziggo internet provider
in the Netherlands.

Mostly copied from [compal_CH7465LG_py](https://github.com/ties/compal_CH7465LG_py)
and [connectbox-prometheus](https://github.com/mbugert/connectbox-prometheus).

## Build and run

Copy and populate config
```sh
cp config.example.yaml config.yaml
```

Start
```sh
make build run
```

## Get metrics

Get exporter internal metrics
```sh
curl 'http://localhost:8080/metrics'
```

Get ConnectBox metrics
```sh
curl 'http://localhost:8080/probe?target=192.168.178.1'
```
