#!/usr/bin/env bash

set -e

CURRENT_PATH=$(cd $(dirname ${BASH_SOURCE[0]}); pwd)
source ${CURRENT_PATH}/x.sh
PROJECT_PATH=$(dirname "${CURRENT_PATH}")
BUILD_PATH=${CURRENT_PATH}/build_solo
BUILD_PATH_TEMP=${BUILD_PATH}/temp

BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function print_red() {
    printf "${RED}%s${NC}\n" "$1"
}

function start() {
  print_blue "===> Start solo axiom-ledger"
  rm -rf "${BUILD_PATH}" && mkdir ${BUILD_PATH} && mkdir ${BUILD_PATH_TEMP}
  export AXIOM_LEDGER_CONSENSUS_TYPE=solo
  ${PROJECT_PATH}/bin/axiom-ledger cluster generate-default --target ${BUILD_PATH_TEMP} --force
  cp -rf ${BUILD_PATH_TEMP}/node1/* ${BUILD_PATH}/
  cp -r ${CURRENT_PATH}/package/* ${BUILD_PATH}/
  cp -f ${PROJECT_PATH}/bin/axiom-ledger ${BUILD_PATH}/tools/bin/
  rm -rf cp -rf ${BUILD_PATH_TEMP}
  echo "AXIOM_LEDGER_KEYSTORE_PASSWORD=2023@axiomesh" >> ${BUILD_PATH}/.env
  cat "${CURRENT_PATH}/package/.env" >> ${BUILD_PATH}/.env
  ${BUILD_PATH}/axiom-ledger start
}

start
