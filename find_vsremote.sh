#!/bin/bash

# Find .vscode and .vscode-server directories at depth 1
initial_results=$(fd --hidden '^\.(vscode|vscode-server)$' "$HOME" -d 1)

# Iterate over each found directory to search for remote-cli directories
while IFS= read -r dir; do
    # Search for 'remote-cli' directories within each found directory
    fd 'remote-cli' "$dir" --type d --max-depth 4 | while IFS= read -r remote_cli_dir; do
        # For each remote-cli directory, search for the 'code' binary
        fd '^code$' "$remote_cli_dir" --type f --exec echo "Found code binary at: {}"
    done
done <<< "$initial_results"
