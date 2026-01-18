#!/bin/bash

# Script to add volume mounts to Kubernetes YAML file for message logging
# This script adds hostPath volume mounts to all service deployments
# so that logs written to /var/log/arpc-messages inside containers
# appear in /users/aruj/hotel-reservation-arpc/logs on the host machine

set -e

YAML_FILE="/users/aruj/hotel-reservation-arpc/hotel_reservation.yaml"
BACKUP_FILE="/users/aruj/hotel-reservation-arpc/hotel_reservation.yaml.bak"
LOG_DIR="/users/aruj/hotel-reservation-arpc/logs"

# Services that need message logging (aRPC services, not mongodb/memcached)
SERVICES="user search reservation recommendation rate profile geo frontend"

# Create backup
cp "$YAML_FILE" "$BACKUP_FILE"
echo "Created backup: $BACKUP_FILE"

# Create logs directory if it doesn't exist
mkdir -p "$LOG_DIR"
echo "Created logs directory: $LOG_DIR"

# Use Python to modify the YAML file properly
python3 << 'PYTHON_SCRIPT'
import yaml
import sys

yaml_file = "/users/aruj/hotel-reservation-arpc/hotel_reservation.yaml"
log_dir = "/users/aruj/hotel-reservation-arpc/logs"
services = ["user", "search", "reservation", "recommendation", "rate", "profile", "geo", "frontend"]

# Load all documents from the YAML file
with open(yaml_file, 'r') as f:
    docs = list(yaml.safe_load_all(f))

modified_count = 0

for doc in docs:
    if doc is None:
        continue

    # Check if this is a Deployment for one of our services
    if doc.get('kind') == 'Deployment':
        app_label = doc.get('metadata', {}).get('labels', {}).get('app', '')

        if app_label in services:
            spec = doc.get('spec', {}).get('template', {}).get('spec', {})
            containers = spec.get('containers', [])

            if containers:
                container = containers[0]

                # Add volumeMounts to container if not exists
                if 'volumeMounts' not in container:
                    container['volumeMounts'] = []

                # Check if message-logs mount already exists
                has_log_mount = any(vm.get('name') == 'message-logs' for vm in container['volumeMounts'])

                if not has_log_mount:
                    container['volumeMounts'].append({
                        'name': 'message-logs',
                        'mountPath': '/var/log/arpc-messages'
                    })

                    # Add volumes to spec if not exists
                    if 'volumes' not in spec:
                        spec['volumes'] = []

                    # Check if message-logs volume already exists
                    has_log_volume = any(v.get('name') == 'message-logs' for v in spec['volumes'])

                    if not has_log_volume:
                        spec['volumes'].append({
                            'name': 'message-logs',
                            'hostPath': {
                                'path': log_dir,
                                'type': 'DirectoryOrCreate'
                            }
                        })

                    modified_count += 1
                    print(f"Updated: {app_label}")
                else:
                    print(f"Skipped (already has mount): {app_label}")

# Write back to file
with open(yaml_file, 'w') as f:
    yaml.dump_all(docs, f, default_flow_style=False, sort_keys=False)

print(f"\nModified {modified_count} deployments")
PYTHON_SCRIPT

echo ""
echo "Volume mounts added successfully!"
echo "Logs will be written to: $LOG_DIR"
echo ""
echo "To revert changes, run: cp $BACKUP_FILE $YAML_FILE"
