#! /bin/bash
set -e

base_dir=$(cd $(dirname ${BASH_SOURCE[0]}); pwd)
env_file=${base_dir}/.env
if [ -f ${env_file} ]; then
  export AXIOM_LEDGER_ENV_FILE=${env_file}
fi
export AXIOM_LEDGER_PATH=${base_dir}

${base_dir}/tools/control.sh status