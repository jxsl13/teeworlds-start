#!/bin/bash

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )/../ranking/"

RANKING_FILE=ranking.db
RANKING_PATH="$(realpath $SCRIPT_DIR)"
RANKING_FILE_PATH="${RANKING_PATH}/${RANKING_FILE}"

BACKUP_PATH="${RANKING_PATH}/backups"
BACKUP_FILE_PATH="${BACKUP_PATH}/$(date "+%Y.%m.%d-%H.%M.%S")_${RANKING_FILE}"

# createbackup directory if it doesn't exist
mkdir -p ${BACKUP_PATH}

cp ${RANKING_FILE_PATH} ${BACKUP_FILE_PATH}