# ConnectBox Exporter

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
