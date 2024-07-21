<p align="center">
  <img src="docs/logo.png" alt="DMap-Zero Logo" width="200"/>
</p>

<h1 align="center">DMap-Zero</h1>
<p align="center">
  Distributed Multi-Agent Platform for Zero-Trust Security
</p>
<p align="center">
  <img src="https://github.com/andreepyro/dmap-zero/actions/workflows/go.yml/badge.svg" alt="Build Status">
  <!-- <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
  </a> -->
</p>

## Prerequisites

- Linux OS (developed and tested on Ubuntu 24.04, [install instructions](https://ubuntu.com/tutorials/install-ubuntu-desktop))
- `Go > 1.23` ([install instructions](https://go.dev/doc/install))
- Docker Engine ([install instructions](https://docs.docker.com/engine/install/))
- OpenZiti CLI ([install instructions](https://openziti.io/docs/downloads/?os=Linux))

## Running Platform Locally

There are many more arguments and settings you can run the scripts with. For more context, look at the scripts.

Build `agent-controller` and `poc-module` docker images, and `dmapz-agent` binary:
```bash
make build
```

Start Main Node infrastructure with agent controller:
```bash
make run controller env="OPENZITI_PASSWORD=<password> GRAFANA_PASSWORD=<password> AGENT_CONTROLLER_CREDENTIALS=<username>:<password>"
```

Start Node infrastructure with agent:
```bash
make run agent env="OPENZITI_PASSWORD=<password> AGENT_CONTROLLER_CREDENTIALS=<username>:<password> CONTROLLER_ADDRESS=<addr> NODE_NUM=<1-99>"
```

## Deployment on VM

Build the project locally:
```bash
make build
```

Prepare the deployment files:
```bash
./scripts/prepare_deployment.sh
```

### Main Node setup
Copy `main_node.tar.gz` to the Main Node.
```bash
scp ./deployment/main_node.tar.gz user@192.168.0.5:~/main_node.tar.gz
```

Setup and start the Main Node:
```bash
ssh user@192.168.0.5
tar -xvzf main_node.tar.gz
./scripts/main_node_setup.sh
OPENZITI_PASSWORD=notsosecure GRAFANA_PASSWORD=notsosecure AGENT_CONTROLLER_CREDENTIALS=username:password ADVERTISED_IP=192.168.0.5.sslip.io ./scripts/main_node_start.sh
```

### Node setup
Copy `node.tar.gz` to the Node.
```bash
scp ./deployment/node.tar.gz user@192.168.0.6:~/node.tar.gz
```

Setup and start the Node:
```bash
ssh user@192.168.0.6
tar -xvzf node.tar.gz
./scripts/node_setup.sh
OPENZITI_PASSWORD=notsosecure AGENT_CONTROLLER_CREDENTIALS=username:password CONTROLLER_ADDRESS=192.168.0.5.sslip.io NODE_NUM=1 ./scripts/node_start.sh
```

## Management console

Management console can be accessed on: https://localhost:6969/ (or different IP depending on your deployment)

## Logs & Metrics

Logs and metrics can be accessed in Grafana running on: http://localhost:3000/ (or different IP depending on your deployment)
