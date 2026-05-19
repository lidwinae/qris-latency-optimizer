#!/usr/bin/env sh
set -eu

K6_OUT_URL="${K6_OUT_URL:-influxdb=http://localhost:8086/k6}"

echo "Running optimized scenario -> ${K6_OUT_URL}"
k6 run --out "${K6_OUT_URL}" tests_script/06-optimized-dashboard.js

echo "Running non-optimized scenario -> ${K6_OUT_URL}"
k6 run --out "${K6_OUT_URL}" tests_script/07-non-optimized-dashboard.js
