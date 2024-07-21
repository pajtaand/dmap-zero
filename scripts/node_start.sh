#!/bin/bash

# Define default variables
DEFAULT_COMPOSE_FILE="./docker/node-compose.yaml"
DEFAULT_ADVERTISED_IP="$(hostname -I | awk '{print $1}').sslip.io"
DEFAULT_HEALTHCHECK_TIMEOUT=30

# Set variables using defaults if not already defined
COMPOSE_FILE="${COMPOSE_FILE:-$DEFAULT_COMPOSE_FILE}"
ADVERTISED_IP="${ADVERTISED_IP:-$DEFAULT_ADVERTISED_IP}"
HEALTHCHECK_TIMEOUT="${HEALTHCHECK_TIMEOUT:-$DEFAULT_HEALTHCHECK_TIMEOUT}"

# IP check
[ "$ADVERTISED_IP" = ".sslip.io" ] && { echo "Error: No IP address found."; exit 1; }

# Check if NODE_NUM is set and is a number
if [[ -z "$NODE_NUM" ]]; then
    NODE_NUM=1
elif ! [[ "$NODE_NUM" =~ ^[0-9]+$ ]]; then
    echo "NODE_NUM is not a valid number."
    exit 1
fi

# Check if NODE_NUM is between 0 and 99
if (( NODE_NUM < 0 || NODE_NUM > 99 )); then
    echo "NODE_NUM must be between 0 and 99."
    exit 1
fi

# Check that OPENZITI_PASSWORD is provided
if [ -z "$OPENZITI_PASSWORD" ]; then
  echo "Error: Environment variable OPENZITI_PASSWORD is not set."
  exit 1
fi

# Check that CONTROLLER_ADDRESS is provided
if [ -z "$CONTROLLER_ADDRESS" ]; then
  echo "Error: Environment variable CONTROLLER_ADDRESS is not set."
  exit 1
fi

# Check that AGENT_CONTROLLER_CREDENTIALS is provided
if [ -z "$AGENT_CONTROLLER_CREDENTIALS" ]; then
  echo "Error: Environment variable AGENT_CONTROLLER_CREDENTIALS is not set."
  exit 1
fi

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
AGENT_JWT="$AGENT_JWT" \
  ROUTER_PORT="$(( NODE_NUM + 3022 ))" \
  CONTROLLER_ADDRESS="$CONTROLLER_ADDRESS" \
  NODE_NUM="$NODE_NUM" \
  ADVERTISED_IP="$ADVERTISED_IP" \
  ROUTER_ENROLLMENT_JWT="$router_jwt" \
  HOST_JWT="$host_jwt" \
  PROMTAIL_JWT="$promtail_jwt" \
  docker compose -f "$COMPOSE_FILE" -p "$NODE_NUM" down -v

# Setup OpenZiti
echo "Logging in to the OpenZiti API..."
for i in {1..5}; do
  ziti edge login "$CONTROLLER_ADDRESS:1280" -y -u admin -p "$OPENZITI_PASSWORD" && break
  sleep 5
  echo "Trying again..."
done

# Clean all objects to prevent errors
echo "Cleaning up all required entities"
ziti edge delete edge-router "er-node-$NODE_NUM" 
ziti edge delete identity "promtail-$NODE_NUM" "node-host-$NODE_NUM"

# Create policies
echo "Setting up OpenZiti"
# Create identity for edge router
ziti edge create edge-router "er-node-$NODE_NUM" -a node-router-role -o /tmp/er.jwt
router_jwt="$(cat /tmp/er.jwt)"
# Create identity for promtail ziti tunneler
ziti edge create identity "promtail-$NODE_NUM" -a promtail-role -o /tmp/promtail.jwt
promtail_jwt="$(cat /tmp/promtail.jwt)"
# Create identity for ziti host
ziti edge create identity "node-host-$NODE_NUM" -a node-role -o /tmp/node-host.jwt
host_jwt="$(cat /tmp/node-host.jwt)"

# Create agent
echo "Creating new agent..."
CREATE_AGENT_RESPONSE=$(curl -s -k -u "$AGENT_CONTROLLER_CREDENTIALS" -X POST \
  -H "Content-Type: application/json" \
  -d "{\"Name\":\"agent-$NODE_NUM\",\"Configuration\":{\"config\":\"value\"}}" \
  https://$CONTROLLER_ADDRESS:6969/api/v1/agent)

# Extract the ID from the response
AGENT_ID=$(echo "$CREATE_AGENT_RESPONSE" | sed -n 's/.*"ID":"\([^"]*\)".*/\1/p')

# Check if AGENT_ID was extracted successfully
if [ -z "$AGENT_ID" ]; then
  echo "Failed to extract agent ID."
  exit 1
fi

echo "Agent ID: $AGENT_ID"

# Enrolle agent
echo "Enrolling newly created agent..."
ENROLLMENT_RESPONSE=$(curl -s -k -u "$AGENT_CONTROLLER_CREDENTIALS" -X POST \
  https://$CONTROLLER_ADDRESS:6969/api/v1/agent/$AGENT_ID/enrollment)

# Extract the JWT from the response
AGENT_JWT=$(echo "$ENROLLMENT_RESPONSE" | sed -n 's/.*"JWT":"\([^"]*\)".*/\1/p')

# Check if JWT was extracted successfully
if [ -z "$AGENT_JWT" ]; then
  echo "Failed to extract JWT."
  exit 1
fi

# Run docker containers
echo "Running Ziti tunneler and promtail containers..."
AGENT_JWT="$AGENT_JWT" \
  ROUTER_PORT="$(( NODE_NUM + 3022 ))" \
  CONTROLLER_ADDRESS="$CONTROLLER_ADDRESS" \
  NODE_NUM="$NODE_NUM" \
  ADVERTISED_IP="$ADVERTISED_IP" \
  ROUTER_ENROLLMENT_JWT="$router_jwt" \
  HOST_JWT="$host_jwt" \
  PROMTAIL_JWT="$promtail_jwt" \
  docker compose -f "$COMPOSE_FILE" -p "$NODE_NUM" up -d -V --force-recreate ziti-router-agent ziti-tunnel-agent promtail-agent node-exporter-agent ziti-host-agent dmapz-agent

echo "Setup complete!"
