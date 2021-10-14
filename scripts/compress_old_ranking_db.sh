#!/bin/bash
# in case you have got demos enabled, you can add this script to your crontab in order to compress old demos.

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )/../ranking/backups/"

DB_PATH="$(realpath $SCRIPT_DIR)"

ZIP_PATH="${DB_PATH}"
ZIP_FILE_PATH="${ZIP_PATH}/$(date +%Y-%m-%d)_ranking.tar.gz"

# createbackup directory if it doesn't exist
mkdir -p ${ZIP_PATH}

# find demos of 30 days ago.
find ${DB_PATH} -type f -name '*.dv' -mtime +30 -print0 | tar -czvf ${ZIP_FILE_PATH} --remove-files --null -T -

