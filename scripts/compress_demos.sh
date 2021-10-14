#!/bin/bash
# in case you have got demos enabled, you can add this script to your crontab in order to compress old demos.

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )/../demos/"

DEMO_PATH="$(realpath $SCRIPT_DIR)"
AUTO_DEMO_PATH="${DEMO_PATH}/auto/"

ZIP_PATH="${DEMO_PATH}/auto/old/"
ZIP_FILE_PATH="${ZIP_PATH}/$(date -d "$(date) - 30 days" +%Y-%m-%d)_demos.tar.gz"


# createbackup directory if it doesn't exist
mkdir -p ${ZIP_PATH}

# find demos of 30 days ago.
find ${AUTO_DEMO_PATH} -type f -name '*.demo' -mtime +30 -print0 | tar -czvf ${ZIP_FILE_PATH} --remove-files --null -T -

