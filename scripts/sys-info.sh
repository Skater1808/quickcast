#!/usr/bin/env bash

# @raycast.schemaVersion 1
# @raycast.title System Resource Overview
# @raycast.mode fullOutput
# @raycast.packageName System
# @raycast.icon 📊
# @raycast.description Show RAM, CPU, and Disk metrics

echo "=== System Overview ==="
uname -a
echo ""
echo "=== Memory Usage ==="
free -h 2>/dev/null || vm_stat
echo ""
echo "=== Disk Space ==="
df -h /
