#!/bin/bash

GOOS=linux go build -o ./rabbitmq-auto-scaler .

docker build -t idobry/rabbitmq-auto-scaler:$1 .

docker push idobry/rabbitmq-auto-scaler:$1

docker tag idobry/rabbitmq-auto-scaler:$1 idobry/rabbitmq-auto-scaler:latest

docker push idobry/rabbitmq-auto-scaler:latest

k delete deploy rabbitmq-auto-scaler ; k apply -f kubernetes/deploy.yaml


