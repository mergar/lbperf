#!/bin/sh
#curl -s http://localhost:8080/api/v1/graph/1700000000 | jq

curl -s -X POST http://172.16.0.92:8080/api/v1/compare/ \
  -H "Content-Type: application/json" \
  -d '{"ids": [1782905422, 1782905421]}' | jq

#curl -s -X POST http://localhost:8080/api/v1/compare/ \
#  -H "Content-Type: application/json" \
#  -d '{"ids": [1700000000, 1700000001, 1700000002]}' | jq


