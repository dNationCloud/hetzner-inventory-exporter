# Hetzner Cloud inventory exporter

Prometheus exporter of Hetzner Cloud inventory


## Build

- Using docker

_Requires docker_

```bash
make build
```

- Locally

_Requires go build environment_

```bash
go build
```

## Run

First, create `config.yaml` file with following content:

```yaml
---
targets:
  - name: my-project-1
    apiKey: 12345678654321234567876543223456787654
  - name: my-project-2
    apiKey: 09876543212456709867234567890987656789
  - name: my-project-3
    apiKey: abcdefg5384572872131237182738174849575
```

- Run using docker

```bash
docker run -p 9112:9112 -v $(pwd)/config.yaml:/config.yaml -ti hetzner-inventory-exporter:1.0.0
```

- Run locally

```bash
./hetzner-inventory-exporter
```

## Exported metrics

| Metric name             |     Type | Description |
|-------------------------|----------|-------------|
| hetzner_exporter_last_scrape_error | Gauge | Number of errors in last scrape round |
| hetzner_exporter_scrapes_total | Counter | Total number of scrapes |
| hetzner_inventory_image_count | Gauge | Number of images (snapshots,backups) in project |
| hetzner_inventory_image_size | Gauge | Cumulative size of all images in project |
| hetzner_inventory_load_balancers | Gauge | Number of load balancers in project |
| hetzner_inventory_network_count | Gauge | Number of networks in project |
| hetzner_inventory_server_count | Gauge | Number of servers in project |
| hetzner_inventory_volume_count | Gauge | Number of volumes in project |
| hetzner_inventory_volume_size | Gauge | Cumulative size of all volumes in project |

## Example exporter output

```
curl --silent localhost:9112/metrics | grep ^hetzner

hetzner_exporter_last_scrape_error 0
hetzner_exporter_scrapes_total 2
hetzner_inventory_exporter_build_info{branch="",goversion="go1.16.8",revision="",version=""} 1
hetzner_inventory_image_count{project="my-project-1"} 2
hetzner_inventory_image_count{project="my-project-2"} 1
hetzner_inventory_image_count{project="my-project-3"} 1
hetzner_inventory_image_size{project="my-project-1"} 2.327502489089966
hetzner_inventory_image_size{project="my-project-2"} 0.7621023654937744
hetzner_inventory_image_size{project="my-project-3"} 0.7770731449127197
hetzner_inventory_load_balancers{project="my-project-1"} 0
hetzner_inventory_load_balancers{project="my-project-2"} 2
hetzner_inventory_load_balancers{project="my-project-3"} 3
hetzner_inventory_network_count{project="my-project-1"} 1
hetzner_inventory_network_count{project="my-project-2"} 1
hetzner_inventory_network_count{project="my-project-3"} 1
hetzner_inventory_server_count{project="my-project-1"} 0
hetzner_inventory_server_count{project="my-project-2"} 9
hetzner_inventory_server_count{project="my-project-3"} 7
hetzner_inventory_volume_count{project="my-project-1"} 0
hetzner_inventory_volume_count{project="my-project-2"} 2
hetzner_inventory_volume_count{project="my-project-3"} 2
hetzner_inventory_volume_size{project="my-project-1"} 0
hetzner_inventory_volume_size{project="my-project-2"} 120
hetzner_inventory_volume_size{project="my-project-3"} 120
```