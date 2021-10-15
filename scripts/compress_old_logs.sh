#!/bin/bash
# in case you have got logs enabled, you can add this script to your crontab in order to compress old logs.

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )/../logs/"
LOGS_PATH="$(realpath $SCRIPT_DIR)"

ZIP_PATH="${LOGS_PATH}/old/"
ZIP_FILE_PATH="${ZIP_PATH}/$(date -d "$(date) - 30 days" +%Y-%m-%d)_logs.tar.gz"

# create backup directory if it doesn't exist
mkdir -p ${ZIP_PATH}

# find demos of 30 days ago. .txt or .log files
find ${LOGS_PATH} -type f \( -name \*.log -o -name \*.txt \) -mtime +30 -print0 | tar -czvf ${ZIP_FILE_PATH} --remove-files --null -T -

