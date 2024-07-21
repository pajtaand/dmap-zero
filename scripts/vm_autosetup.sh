#!/bin/bash

# Check if the script is run with sudo privileges
if [ "$EUID" -ne 0 ]; then
  echo "Please run this script with sudo privileges."
  exit 1
fi

# Check if NUM is provided as an argument
if [ -z "$1" ]; then
  echo "Usage: $0 NUM (number between 1 and 99)"
  exit 1
fi

NUM="$1"

# Validate that NUM is an integer between 1 and 99
if ! [[ "$NUM" =~ ^[0-9]+$ ]]; then
  echo "Error: NUM must be an integer."
  exit 1
fi

if [ "$NUM" -lt 1 ] || [ "$NUM" -gt 99 ]; then
  echo "Error: NUM must be between 1 and 99."
  exit 1
fi

NUM_PLUS_10=$((NUM + 10))

# Replace current IP in /etc/netplan/50-cloud-init.yaml using sed
sed -i "s/192\.168\.0\.11\/24/192\.168\.0\.${NUM_PLUS_10}\/24/g" /etc/netplan/50-cloud-init.yaml

# Apply the netplan configuration
netplan apply

# Request new IP using dhcpcd
dhcpcd -r 10.0.13.$NUM_PLUS_10

# Add Agent startup to .bashrc file
echo './scripts/cleanup.sh' >> /home/user/.bashrc
echo "OPENZITI_PASSWORD=notsosecure AGENT_CONTROLLER_CREDENTIALS=username:password NODE_NUM=$NUM CONTROLLER_ADDRESS=192.168.0.5.sslip.io ADVERTISED_IP=192.168.0.$NUM_PLUS_10.sslip.io ./scripts/node_start.sh" >> /home/user/.bashrc

# Reboot the machine
reboot
