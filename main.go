package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/streadway/amqp"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

const (
	notFound = "not found"
)

var (
	rabbitmqUrl  = fmt.Sprintf("amqp://%s:%s@%s", os.Getenv("RABBITMQ_USER"), os.Getenv("RABBITMQ_PASSWORD"), os.Getenv("RABBITMQ_URL"))
	minPods, _   = strconv.Atoi(os.Getenv("MIN_POD"))
	maxPods, _   = strconv.Atoi(os.Getenv("MAX_POD"))
	msgPerPod, _ = strconv.Atoi(os.Getenv("MSG_PER_POD"))
	interval, _ := strconv.Atoi(os.Getenv("INTERVAL"))
)

type Queues struct {
	Arguments struct {
	} `json:"arguments"`
	AckRequired    bool `json:"ack_required"`
	ChannelDetails struct {
		ConnectionName string `json:"connection_name"`
		Name           string `json:"name"`
		Node           string `json:"node"`
		Number         int    `json:"number"`
		PeerHost       string `json:"peer_host"`
		PeerPort       int    `json:"peer_port"`
		User           string `json:"user"`
	} `json:"channel_details"`
	ConsumerTag   string `json:"consumer_tag"`
	Exclusive     bool   `json:"exclusive"`
	PrefetchCount int    `json:"prefetch_count"`
	Queue         struct {
		Name  string `json:"name"`
		Vhost string `json:"vhost"`
	} `json:"queue"`
}

func main() {
	fmt.Printf("\n\n*********** Connecting to Rabbimq at %s ***********\n", rabbitmqUrl)
	conn, _ := amqp.Dial(rabbitmqUrl)
	defer conn.Close()

	// create a channel
	ch, _ := conn.Channel()
	defer ch.Close()

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		queues := GetQueues()
		for _, q := range queues {
			Run(q, ch, clientset)
		}
		time.Sleep(time.Duration(int(time.Second) * int(interval)))
	}
}

func Run(q Queues, ch *amqp.Channel, clientset *kubernetes.Clientset) {
	// get ready queue msgs count
	queue, err := ch.QueueInspect(q.Queue.Name)
	if err != nil {
		fmt.Println(err.Error())
	}
	numToScale := GetScaleCount(queue.Messages)

	ScaleDeployment(clientset, "default", GetDeploymentName(clientset, q.ChannelDetails.PeerHost), numToScale)
}

func GetQueues() []Queues {
	manager := os.Getenv("RABBITMQ_MANAGMENT_URL") + "/api/consumers/"
	client := &http.Client{}
	req, err := http.NewRequest("GET", manager, nil)
	if err != nil {
		fmt.Println(err.Error())
	}
	req.SetBasicAuth(os.Getenv("RABBITMQ_USER"), os.Getenv("RABBITMQ_PASSWORD"))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}

	value := make([]Queues, 0)
	json.NewDecoder(resp.Body).Decode(&value)
	return value
}

func GetDeploymentName(clientset *kubernetes.Clientset, podIP string) string {
	pods, err := clientset.CoreV1().Pods(os.Getenv("NAMESPACE")).List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		return ""
	}
	for _, pod := range pods.Items {
		if pod.Status.PodIP == podIP {
			delimiter := "-"
			return strings.Join(strings.Split(pod.Name, delimiter)[:(strings.Count(pod.Name, delimiter)-1)], delimiter)
		}
	}
	return notFound
}

func ScaleDeployment(clientset *kubernetes.Clientset, namespace string, deploymentName string, numToScale int) {
	if deploymentName != notFound {
		deploymentsClient := clientset.AppsV1().Deployments(namespace)
		deployment, err := deploymentsClient.Get(deploymentName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Deployment not found\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			fmt.Println(err.Error())
		}
		if *deployment.Spec.Replicas == int32(numToScale) {
			fmt.Printf("No need to scale %s, current replicas=%d\n", deploymentName, *deployment.Spec.Replicas)
		} else {
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				// Retrieve the latest version of Deployment before attempting update
				// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
				result, getErr := deploymentsClient.Get(deploymentName, metav1.GetOptions{})
				if getErr != nil {
					panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
				}

				result.Spec.Replicas = int32Ptr(int32(numToScale))
				_, updateErr := deploymentsClient.Update(result)
				return updateErr
			})

			if retryErr != nil {
				panic(fmt.Errorf("Update failed: %v\n", retryErr))
			}
			fmt.Printf("Deploymnet %s scaled from %d to %d\n", deploymentName, *deployment.Spec.Replicas, numToScale)
		}
	}
}

func GetScaleCount(currentMsgInQueue int) int {
	if currentMsgInQueue >= (msgPerPod * maxPods) {
		return maxPods
	}
	if currentMsgInQueue <= (msgPerPod * minPods) {
		return minPods
	}
	res := int(math.Ceil(float64(currentMsgInQueue) / float64(msgPerPod)))
	return res
}

func int32Ptr(i int32) *int32 { return &i }
