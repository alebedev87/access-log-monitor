#! /bin/bash

paths=("/" "/here" "/there")
len=${#paths[*]}

while true; do
    idx=$(( ${RANDOM} % ${len} ))
    path="${paths[${idx}]}"
    curl -s http://localhost:8080"${path}" > /dev/null
    if [ -n "${SLEEPTIME}" ]; then
        sleep "${SLEEPTIME}"
    fi
done
