#!/bin/bash

set -e

DIR="$(cd "$(dirname "${0}")" && pwd)"
cd "${DIR}"

declare -A MAP=( \
  ['Zap']=':zap: zap' \
  ['Zap.Sugar']=':zap: zap (sugared)' \
  ['stdlib.Println']='standard library' \
  ['sirupsen/logrus']='logrus' \
  ['go-kit/kit/log']='go-kit' \
  ['inconshreveable/log15']='log15' \
  ['apex/log']='apex/log' \
  ['go.pedge.io/lion']='lion' \
)
declare -a KEYS=( \
  'Zap' \
  'Zap.Sugar' \
  'stdlib.Println' \
  'sirupsen/logrus' \
  'go-kit/kit/log' \
  'inconshreveable/log15' \
  'apex/log' \
  'go.pedge.io/lion' \
)

benchmarks() {
  benchmark_adding_fields
  echo
  benchmark_accumulated_context
  echo
  benchmark_without_fields
}

benchmark_adding_fields() {
  echo 'Log a message and 10 fields:'
  echo
  benchmark_rows 'BenchmarkAddingFields'
}

benchmark_accumulated_context() {
  echo 'Log a message with a logger that already has 10 fields of context:'
  echo
  benchmark_rows 'BenchmarkAccumulatedContext'
}

benchmark_without_fields() {
  echo 'Log a static string, without any context or `printf`-style templating:'
  echo
  benchmark_rows 'BenchmarkWithoutFields'
}

benchmark_rows() {
  echo '| Library | Time | Bytes Allocated | Objects Allocated |'
  echo '| :--- | :---: | :---: | :---: |'
  for key in ${!KEYS[@]}; do
    benchmark_row "${1}" "${KEYS[${key}]}"
  done
}

benchmark_row() {
  if cat "${TMP}" | grep "${1}/${2}" >/dev/null; then
    echo "|${MAP[${2}]}|$(cat "${TMP}" | grep "${1}/${2}-" | cut -f 3-5 -d '|')|"
  fi
}

TMP="$(mktemp)"
trap "rm -f ${TMP}" EXIT
make bench | tr '\t' '|' > "${TMP}"
README_TMP="$(mktemp)"
trap "rm -f ${README_TMP}" EXIT
head -$(cat README.md | grep -n '### Benchmarks' | cut -f 1 -d :) README.md >> "${README_TMP}"
echo >> "${README_TMP}"
benchmarks >> "${README_TMP}"
mv -f "${README_TMP}" README.md
