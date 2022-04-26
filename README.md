Add messages to a rabbitMQ queue; remove them from it
We will autoscale components based on the Prometheus metric exported from rabbitMQ, tracking the number of messages sitting in the queue

## Usage

Add ten messages to the "autoscaling" queue
```
git clone https://github.com/adrianparks/thumper.git
kubectl port-forward sts/nfsaas-rabbitmq 5672 &
kubectl port-forward sts/nfsaas-rabbitmq 15672 &

export RABBITMQ_DEFAULT_USER=$(kubectl exec -it nfsaas-rabbitmq-0 -c nfsaas-rabbitmq -- env | grep RABBITMQ_DEFAULT_USER | cut -d '=' -f 2 | tr -d '\r')
export RABBITMQ_DEFAULT_PASS=$(kubectl exec -it nfsaas-rabbitmq-0 -c nfsaas-rabbitmq -- env | grep RABBITMQ_DEFAULT_PASS | cut -d '=' -f 2 | tr -d '\r')
cd /thumper
go run send/send.go -n 10
```