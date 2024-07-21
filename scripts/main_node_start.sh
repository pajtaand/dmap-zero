#!/bin/bash

# Define default variables
DEFAULT_COMPOSE_FILE="./docker/main-node-compose.yaml"
DEFAULT_ADVERTISED_IP="$(hostname -I | awk '{print $1}').sslip.io"
DEFAULT_OPENZITI_PASSWORD=$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16)
DEFAULT_GRAFANA_PASSWORD=$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16)
DEFAULT_AGENT_CONTROLLER_CREDENTIALS="admin:$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16)"
DEFAULT_HEALTHCHECK_TIMEOUT=30

# Set variables using defaults if not already defined
COMPOSE_FILE="${COMPOSE_FILE:-$DEFAULT_COMPOSE_FILE}"
ADVERTISED_IP="${ADVERTISED_IP:-$DEFAULT_ADVERTISED_IP}"
OPENZITI_PASSWORD="${OPENZITI_PASSWORD:-$DEFAULT_OPENZITI_PASSWORD}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-$DEFAULT_GRAFANA_PASSWORD}"
AGENT_CONTROLLER_CREDENTIALS="${AGENT_CONTROLLER_CREDENTIALS:-$DEFAULT_AGENT_CONTROLLER_CREDENTIALS}"
HEALTHCHECK_TIMEOUT="${HEALTHCHECK_TIMEOUT:-$DEFAULT_HEALTHCHECK_TIMEOUT}"

# IP check
[ "$ADVERTISED_IP" = ".sslip.io" ] && { echo "Error: No IP address found."; exit 1; }

# Function to check if a command is available
check_command() {
  local cmd=$1
  if command -v "$cmd" >/dev/null 2>&1; then
    echo "$cmd is available."
  else
    echo "$cmd is not available. Please install it and try again."
    exit 1
  fi
}

# Function to wait until a container is healthy
wait_for_container_health() {
  local container_name=$1
  local timeout=${2:-$HEALTHCHECK_TIMEOUT}
  local start_time=$(date +%s)
  local end_time=$((start_time + timeout))
  local container_healthy="false"

  echo "Waiting for container $container_name to be healthy..."

  # Poll health status until healthy or timeout
  while [ "$(date +%s)" -lt $end_time ]; do
    health_status=$(docker inspect --format='{{.State.Health.Status}}' "$container_name")
    
    if [ "$health_status" == "healthy" ]; then
      container_healthy="true"
      echo "$container_name is healthy!"
      break
    else
      echo "Current health status of $container_name: $health_status. Retrying in 5 seconds..."
      sleep 5
    fi
  done

  if [ "$container_healthy" != "true" ]; then
    echo "$container_name did not become healthy within timeout."
    exit 1
  fi
}

# Check if required commands are available
echo "Checking that all required commands are available..."
check_command "docker"
check_command "ziti"

# Remove existing Docker containers and volumes
echo "Removing existing containers and volumes..."
docker compose -f "$COMPOSE_FILE" down -v

# Generate self-signed certificate
mkdir -p /tmp/certs
openssl req -x509 -newkey rsa:4096 -keyout /tmp/certs/ca-key.pem -out /tmp/certs/ca-cert.pem -sha256 -days 365 -nodes -subj "/CN=$ADVERTISED_IP"
chmod +r /tmp/certs/ca-key.pem /tmp/certs/ca-cert.pem

# Run the OpenZiti Controller container
echo "Running OpenZiti Controller  and Grafana containers..."
ADVERTISED_IP="$ADVERTISED_IP" OPENZITI_PASSWORD="$OPENZITI_PASSWORD" docker compose -f "$COMPOSE_FILE" up -d -V --force-recreate ziti-controller grafana prometheus node-exporter

# Wait for the controller container to be healthy
wait_for_container_health "ziti-controller"

# Setup OpenZiti
echo "Logging in to the OpenZiti API..."
for i in {1..5}; do
  ziti edge login "$ADVERTISED_IP:1280" -y -u admin -p "$OPENZITI_PASSWORD" && break
  sleep 5
  echo "Trying again..."
done

# Clean all objects to prevent errors
echo "Cleaning up all required entities"
ziti edge delete service-policy sp1 sp2 sp3 sp4 sp5 sp6 sp7 sp8 sp9 sp10
ziti edge delete edge-router-policy erp1 erp2 erp3 erp4 erp5 erp6
ziti edge delete service-edge-router-policy serp1 serp2 serp3 serp4 serp5
ziti edge delete service service-controller service-agent service-p2p service-loki $(echo service-node-exporter-{1..99})
ziti edge delete edge-router er1
ziti edge delete identity controller loki main-node-tunneler
ziti edge delete config $(echo node-exporter-{1..99}.host.v1)

# Create policies
echo "Setting up OpenZiti"
# '#agent-role' role can call '#control-svc-role' group of services
ziti edge create service-policy sp1 Dial --identity-roles '#agent-role' --service-roles '#control-svc-role'
# '#agent-controller-role' role can provide '#control-svc-role' group of services
ziti edge create service-policy sp2 Bind --identity-roles '#agent-controller-role' --service-roles '#control-svc-role'
# '#agent-controller-role' role can call '#agent-svc-role' group of services
ziti edge create service-policy sp3 Dial --identity-roles '#agent-controller-role' --service-roles '#agent-svc-role'
# '#agent-role' role can provide '#agent-svc-role' group of services
ziti edge create service-policy sp4 Bind --identity-roles '#agent-role' --service-roles '#agent-svc-role'
# '#agent-role' role can call '#p2p-svc-role' group of services
ziti edge create service-policy sp5 Dial --identity-roles '#agent-role' --service-roles '#p2p-svc-role'
# '#agent-role' role can provide '#p2p-svc-role' group of services
ziti edge create service-policy sp6 Bind --identity-roles '#agent-role' --service-roles '#p2p-svc-role'
# Allow `#agent-controller-role` role to access '#main-node-router-role' edge routers
ziti edge create edge-router-policy erp1 --identity-roles '#agent-controller-role' --edge-router-roles '#main-node-router-role'
# Allow `#agent-role` role to access '#node-router-role' edge routers
ziti edge create edge-router-policy erp2 --identity-roles '#agent-role' --edge-router-roles '#node-router-role'
# Allow `#p2p-svc-role` group of services to be available from all edge routers
ziti edge create service-edge-router-policy serp1 --service-roles '#p2p-svc-role' --edge-router-roles '#all'
# Allow '#control-svc-role' group of services to be available from all edge routers
ziti edge create service-edge-router-policy serp2 --service-roles '#control-svc-role' --edge-router-roles '#all'
# Allow '#agent-svc-role' group of services to be available from all edge routers
ziti edge create service-edge-router-policy serp3 --service-roles '#agent-svc-role' --edge-router-roles '#all'
# Create control service
ziti edge create service service-controller -a control-svc-role
# Create agent service
ziti edge create service service-agent -a agent-svc-role
# Create p2p agent service
ziti edge create service service-p2p -a p2p-svc-role
# Create identity for edge router
ziti edge create edge-router er1 -a main-node-router-role -o /tmp/er.jwt
router_jwt="$(cat /tmp/er.jwt)"
# Create identity for agent controller
ziti edge create identity controller --admin -a agent-controller-role -o /tmp/agent_controller.jwt
agent_controller_jwt="$(cat /tmp/agent_controller.jwt)"

# Create identity for OpenZiti tunneler for logging
# '#loki-role' role can provide '#loki-svc-role' group of services
ziti edge create service-policy sp7 Bind --identity-roles '#loki-role' --service-roles '#loki-svc-role'
# '#promtail-role' role can call '#loki-svc-role' group of services
ziti edge create service-policy sp8 Dial --identity-roles '#promtail-role' --service-roles '#loki-svc-role'
# Allow `#loki-role` role to access '#main-node-router-role' edge routers
ziti edge create edge-router-policy erp3 --identity-roles '#loki-role' --edge-router-roles '#main-node-router-role'
# Allow `#promtail-role` role to access '#node-router-role' edge routers
ziti edge create edge-router-policy erp4 --identity-roles '#promtail-role' --edge-router-roles '#node-router-role'
# Allow `#loki-svc-role` group of services to be available from all edge routers
ziti edge create service-edge-router-policy serp4 --service-roles '#loki-svc-role' --edge-router-roles '#all'
# Create loki service
ziti edge create config loki.host.v1 host.v1 '{"protocol":"tcp", "address":"loki", "port":3100}'
ziti edge create service service-loki -a loki-svc-role --configs loki.host.v1
# Create identity for loki
ziti edge create identity loki -a loki-role -o /tmp/loki.jwt
loki_jwt="$(cat /tmp/loki.jwt)"

# Setup OpenZiti tunneler for prometheus
# '#node-role' role can provide '#node-svc-role' group of services
ziti edge create service-policy sp9 Bind --identity-roles '#node-role' --service-roles '#node-svc-role'
# '#main-node-role' role can call '#node-svc-role' group of services
ziti edge create service-policy sp10 Dial --identity-roles '#main-node-role' --service-roles '#node-svc-role'
# Allow `#node-role` role to access '#main-node-router-role' edge routers
ziti edge create edge-router-policy erp5 --identity-roles '#node-role' --edge-router-roles '#node-router-role'
# Allow `#main-node-role` role to access '#main-node-router-role' edge routers
ziti edge create edge-router-policy erp6 --identity-roles '#main-node-role' --edge-router-roles '#main-node-router-role'
# Allow `#node-svc-role` group of services to be available from all edge routers
ziti edge create service-edge-router-policy serp5 --service-roles '#node-svc-role' --edge-router-roles '#all'
# Create identity for main node tunneler
ziti edge create identity main-node-tunneler -a main-node-role -o /tmp/main-tunneler.jwt
main_tunneler_jwt="$(cat /tmp/main-tunneler.jwt)"
# Create node services
for i in {1..99}
do
  ziti edge create config "node-exporter-$i.host.v1" host.v1 "{\"protocol\":\"tcp\", \"address\":\"node-exporter-${i}\", \"port\":9100}"
  ziti edge create service "service-node-exporter-$i" -a node-svc-role --configs "node-exporter-$i.host.v1"
done

# Run the OpenZiti Router container
echo "Running OpenZiti Router container..."
ADVERTISED_IP="$ADVERTISED_IP" \
  ROUTER_ENROLLMENT_JWT="$router_jwt" \
  GRAFANA_PASSWORD="$GRAFANA_PASSWORD" \
  LOKI_JWT="$loki_jwt" \
  TUNNELER_JWT="$main_tunneler_jwt" \
  AGENT_CONTROLLER_JWT="$agent_controller_jwt" \
  AGENT_CONTROLLER_CREDENTIALS="$AGENT_CONTROLLER_CREDENTIALS" \
  docker compose -f "$COMPOSE_FILE" up -d -V --force-recreate ziti-router loki promtail ziti-host dmapz-agent-controller ziti-tunnel-controller

echo "Setup complete!"
