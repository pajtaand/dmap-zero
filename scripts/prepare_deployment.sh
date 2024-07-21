#!/bin/bash

# Create empty deployment folder
rm -rf deployment
mkdir -p deployment

# Create Main Node deployment file
echo "Creating Main Node deployment file"
tar -czvf ./deployment/main_node.tar.gz \
  ./scripts/main_node_setup.sh \
  ./scripts/main_node_start.sh \
  ./scripts/cleanup.sh \
  ./docker/main-node-compose.yaml \
  ./configs/grafana-dashboards \
  ./configs/grafana-datasource.yaml \
  ./configs/loki-config.yaml \
  ./configs/openziti-config.yaml \
  ./configs/prometheus-config.yaml \
  ./configs/promtail-config.yaml \
  ./build/image-agent-controller.tar

# Create Node deployment file
echo "Creating Node deployment file"
tar -czvf ./deployment/node.tar.gz \
  ./scripts/node_setup.sh \
  ./scripts/node_start.sh \
  ./scripts/cleanup.sh \
  ./docker/node-compose.yaml \
  ./configs/promtail-config-agent.yaml \
  ./build/image-agent.tar
