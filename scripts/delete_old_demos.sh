#!/bin/bash
# in case you have got demos enabled, you can add this script to your crontab in order to compress old demos.

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )/../demos/"

DEMO_PATH="$(realpath $SCRIPT_DIR)"
AUTO_DEMO_PATH="${DEMO_PATH}/auto"


# delete demos older than 30 days
find ${AUTO_DEMO_PATH} -name "*.demo" -type f -mtime +30 -exec rm {} \; 