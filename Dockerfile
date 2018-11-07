FROM debian
COPY ./rabbit-hpa /rabbit-hpa
ENTRYPOINT /rabbit-hpa
