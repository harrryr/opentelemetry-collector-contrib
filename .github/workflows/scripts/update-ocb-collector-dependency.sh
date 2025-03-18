#!/bin/bash

LOCAL_GO_MOD_PATH=$1
UPSTREAM_GO_MOD_PATH=$2

# This function reads a go.mod file and extracts OTel dependencies in the format go.opentelemetry.io/* or github.com/open-telemetry/opentelemetry-collector-contrib/* 
# in the first require block
extract_opentelemetry_deps() {
    local go_mod_file="$1" 
    local found_first_require=0
    local deps=""

    # Read go.mod file line by line
    while IFS= read -r line; do
        # Detect the first "require (" block
        if [[ $line == "require (" && $found_first_require -eq 0 ]]; then
            found_first_require=1
            continue
        fi

        # Stop capturing when reaching the closing ")" of the first block
        if [[ $line == ")" && $found_first_require -eq 1 ]]; then
            break
        fi

        # Filter out OpenTelemetry dependencies
        if [[ $found_first_require -eq 1 && ($line == *"go.opentelemetry.io"* || $line == *"github.com/open-telemetry/opentelemetry-collector-contrib"*) ]]; then
            deps+="$line"$'\n'  # Append to variable with newline
        fi
    done < "$go_mod_file"

    echo "$deps"  # Return captured dependencies
}

# Extract OpenTelemetry dependencies from both local and remote go.mod files
local_versions=$(extract_opentelemetry_deps $LOCAL_GO_MOD_PATH)
upstream_versions=$(extract_opentelemetry_deps $UPSTREAM_GO_MOD_PATH)

echo "Local OTel Dependency Versions"
echo $local_versions
echo "Upstream OTel Dependency Versions"
echo $upstream_versions

# Iterate over each remote dependency and check against local
while IFS=" " read -r module version; do
    local_version=$(grep -E "^$module " go.mod | awk '{print $2}')
    
    if [[ ! -z "$local_version" && "$local_version" != "$version" ]]; then
        echo "Updating $module from $local_version to $version"
        sed -i "s@$module $local_version@$module $version@" "$LOCAL_GO_MOD_PATH"
    fi
done <<< "$upstream_versions"

go mod tidy



