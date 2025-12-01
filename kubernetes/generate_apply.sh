#!/bin/bash

# Script to generate Kubernetes deployment YAML with custom feature flags
# Usage: ./generate_apply.sh [--reliable=true|false] [--cc=true|false] [--fc=true|false] [--output=filename]

set -e

# Default values (all features enabled)
ENABLE_RELIABLE="true"
ENABLE_CC="true"
ENABLE_FC="true"
OUTPUT_FILE=""

# Parse command line arguments
for arg in "$@"; do
    case $arg in
        --reliable=*)
            ENABLE_RELIABLE="${arg#*=}"
            shift
            ;;
        --cc=*)
            ENABLE_CC="${arg#*=}"
            shift
            ;;
        --fc=*)
            ENABLE_FC="${arg#*=}"
            shift
            ;;
        --output=*)
            OUTPUT_FILE="${arg#*=}"
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--reliable=true|false] [--cc=true|false] [--fc=true|false] [--output=filename]"
            echo ""
            echo "Options:"
            echo "  --reliable=true|false   Enable/disable reliable delivery (default: true)"
            echo "  --cc=true|false         Enable/disable congestion control (default: true)"
            echo "  --fc=true|false         Enable/disable flow control (default: true)"
            echo "  --output=filename       Output filename (default: auto-generated)"
            echo ""
            echo "Examples:"
            echo "  $0 --reliable=false --cc=false --fc=false"
            echo "  $0 --reliable=true --cc=false --fc=false --output=hotel-reliable.yaml"
            echo "  $0 --cc=true --reliable=false --fc=false"
            exit 0
            ;;
        *)
            echo "Unknown option: $arg"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Determine output filename if not specified
if [ -z "$OUTPUT_FILE" ]; then
    features=""
    if [ "$ENABLE_RELIABLE" = "true" ]; then
        features="${features}reliable-"
    fi
    if [ "$ENABLE_CC" = "true" ]; then
        features="${features}cc-"
    fi
    if [ "$ENABLE_FC" = "true" ]; then
        features="${features}fc-"
    fi
    
    if [ -z "$features" ]; then
        OUTPUT_FILE="hotel_reservation_basic.yaml"
    else
        # Remove trailing dash
        features="${features%-}"
        OUTPUT_FILE="hotel_reservation_${features}.yaml"
    fi
fi

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SOURCE_FILE="${PROJECT_ROOT}/hotel_reservation.yaml"
TARGET_FILE="${SCRIPT_DIR}/${OUTPUT_FILE}"

echo "=========================================="
echo "Generating Hotel Reservation deployment"
echo "=========================================="
echo "Source file: ${SOURCE_FILE}"
echo "Target file: ${TARGET_FILE}"
echo ""
echo "Feature flags:"
echo "  ENABLE_RELIABLE: ${ENABLE_RELIABLE}"
echo "  ENABLE_CC:       ${ENABLE_CC}"
echo "  ENABLE_FC:       ${ENABLE_FC}"
echo "=========================================="
echo ""

# Check if source file exists
if [ ! -f "$SOURCE_FILE" ]; then
    echo "Error: Source file ${SOURCE_FILE} does not exist"
    exit 1
fi

# Function to add environment variables to deployments
add_env_vars() {
    local input_file="$1"
    local output_file="$2"
    
    # Use awk to insert environment variables after existing env section in Deployment resources
    awk -v reliable="$ENABLE_RELIABLE" -v cc="$ENABLE_CC" -v fc="$ENABLE_FC" '
    /kind: Deployment/ { in_deployment=1 }
    /^---$/ { 
        if (in_deployment && in_env) {
            # Flush any remaining env vars
            for (i = 1; i <= env_count; i++) {
                print env_lines[i]
            }
            # Add feature flags
            print indent "- name: ENABLE_RELIABLE"
            print indent "  value: \"" reliable "\""
            print indent "- name: ENABLE_CC"
            print indent "  value: \"" cc "\""
            print indent "- name: ENABLE_FC"
            print indent "  value: \"" fc "\""
            # Reset
            in_env=0
            env_count=0
            delete env_lines
        }
        in_deployment=0
        print
        next
    }

    # When we find an env: section in a deployment
    /^[ \t]*env:/ && in_deployment {
        print
        in_env=1
        next
    }

    # Collect env variable lines
    in_env && /^[ \t]*-[ \t]+name:/ {
        env_lines[++env_count] = $0
        # Get indentation from the first env var
        if (indent == "") {
            match($0, /^[ \t]*/)
            indent = substr($0, RSTART, RLENGTH)
        }
        next
    }

    # Collect value lines
    in_env && /^[ \t]*value:/ {
        env_lines[++env_count] = $0
        next
    }

    # When we exit the env section (non-env line with less or equal indentation)
    in_env && !/^[ \t]*-[ \t]+name:/ && !/^[ \t]*value:/ {
        # Check if this line is less indented (means we exited env section)
        match($0, /^[ \t]*/)
        current_indent = substr($0, RSTART, RLENGTH)
        
        # If current line has same or less indentation as env vars, we exited
        if (length(current_indent) <= length(indent) || /^[ \t]*ports:/ || /^[ \t]*resources:/) {
            # Print all stored env lines
            for (i = 1; i <= env_count; i++) {
                print env_lines[i]
            }
            # Add our new environment variables with proper indentation
            print indent "- name: ENABLE_RELIABLE"
            print indent "  value: \"" reliable "\""
            print indent "- name: ENABLE_CC"
            print indent "  value: \"" cc "\""
            print indent "- name: ENABLE_FC"
            print indent "  value: \"" fc "\""
            # Reset
            in_env=0
            env_count=0
            delete env_lines
            indent = ""
        } else {
            # Still in env section (e.g., multi-line value)
            env_lines[++env_count] = $0
            next
        }
    }

    # Default: just print the line
    { print }
    
    END {
        # Handle case where file ends while in env section
        if (in_env) {
            for (i = 1; i <= env_count; i++) {
                print env_lines[i]
            }
            print indent "- name: ENABLE_RELIABLE"
            print indent "  value: \"" reliable "\""
            print indent "- name: ENABLE_CC"
            print indent "  value: \"" cc "\""
            print indent "- name: ENABLE_FC"
            print indent "  value: \"" fc "\""
        }
    }
    ' "$input_file" > "$output_file"
}

add_env_vars "$SOURCE_FILE" "$TARGET_FILE"

echo "=========================================="
echo "✓ Successfully generated deployment YAML"
echo "=========================================="
echo ""
echo "Output file: ${TARGET_FILE}"
echo ""
echo "To deploy:"
echo "  kubectl apply -f ${TARGET_FILE}"
echo ""
echo "To verify:"
echo "  kubectl get pods"
echo "  kubectl get services"
echo ""
echo "To check environment variables:"
echo "  kubectl exec <pod-name> -- env | grep ENABLE"
echo ""
echo "To clean up:"
echo "  kubectl delete -f ${TARGET_FILE}"
echo ""

