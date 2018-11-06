# Rabbit-hpa
[![HitCount](http://hits.dwyl.io/idobry/Rabbit-hpa.svg)](http://hits.dwyl.io/idobry/Rabbit-hpa) [![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/dwyl/esta/issues)

<img src="./logo.png" width="300">

Rabbit-hpa is a service for auto scaling kubernetes pods based on messages count in Rabbit's queues.

### Prerequisites

* Kubernetes cluster
* Rabbitmq and RabbitmqManagment
* Go 1.9 or later

### Installing

```
git clone https://github.com/idobry/Rabbit-hpa.git .

GOOS=linux go build -o ./rabbit-hpa .

docker build .
```

### Deployment

Rabbit-hpa was built to run inside a kubernetes cluster.
Pay attention to the service environment variables, for example:
```
env:
  - name: NAMESPACE
    value: "default"
  - name: RABBITMQ_MANAGMENT_URL
    value: "http://rabbitmq-rabbitmq-ha:15672"
  - name: RABBITMQ_URL
    value: "rabbitmq-rabbitmq-ha:5672"
  - name: RABBITMQ_USER
    value: "user"
  - name: RABBITMQ_PASSWORD
    value: "user"
  - name: INTERVAL
    value: "30"
  - name: MIN_POD
    value: "2"
  - name: MAX_POD
    value: "10"
  - name: MSG_PER_POD
    value: "5"
```
Deploy:
```
kubectl apply -f ./kubernetes
```

## Contributing

Code contributions are very welcome. If you are interested in helping make chartmuseumui great then feel free!

## Versioning

We use [SemVer](http://semver.org/) for versioning.

## Authors

* **Ido Braunstain** - *Initial work*

See also the list of [contributors](https://github.com/your/project/contributors) who participated in this project.
