#!/bin/bash

GOOS=linux go build -o ./rabbit-hpa .

docker build -t idobry/rabbit-hpa:$1 .

docker push idobry/rabbit-hpa:$1

docker tag idobry/rabbit-hpa:$1 idobry/rabbit-hpa:latest

docker push idobry/rabbit-hpa:latest


