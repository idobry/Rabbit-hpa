FROM debian
COPY ./rabbitmq-auto-scaler /rabbitmq-auto-scaler
ENTRYPOINT /rabbitmq-auto-scaler