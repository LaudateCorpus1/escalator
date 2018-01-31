package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"

	"github.com/atlassian/escalator/pkg/controller"
	"github.com/atlassian/escalator/pkg/k8s"
	"gopkg.in/alecthomas/kingpin.v2"

	log "github.com/sirupsen/logrus"
)

var (
	// loglevel = kingpin.Flag("loglevel", "Verbose mode.").Short('v').Default("INFO").String()
	addr               = kingpin.Flag("address", "Address to listen to for /metrics").Default(":8080").String()
	scanInterval       = kingpin.Flag("scaninterval", "How often cluster is reevaluated for scale up or down").Default("60s").Duration()
	kubeConfigFile     = kingpin.Flag("kubeconfig", "Kubeconfig file location").String()
	customerConfigFile = kingpin.Flag("nodegroups", "Config file for customers nodegroups").Required().String()
)

func main() {
	kingpin.Parse()

	// if the kubeConfigFile is in the cmdline args then use the out of cluster config
	var k8sClient kubernetes.Interface
	if kubeConfigFile != nil && len(*kubeConfigFile) > 0 {
		log.Infoln("Using out of cluster config")
		k8sClient = k8s.NewOutOfClusterClient(*kubeConfigFile)
	} else {
		log.Infoln("Using in cluster config")
		k8sClient = k8s.NewInClusterClient()
	}

	// customerConfigFile is required by kingpin. Won't get to here if it's not defined
	configFile, err := os.Open(*customerConfigFile)
	if err != nil {
		log.Fatalf("Failed to open configFile: %v", err)
	}
	customers, err := controller.UnmarshalNodeGroupsConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to decode configFile: %v", err)
	}

	// endpoints
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(*addr, nil)

	opts := &controller.Opts{
		Addr:         *addr,
		ScanInterval: *scanInterval,
		K8SClient:    k8sClient,
		Customers:    customers,
	}

	// signal channel waits for interrupt
	signalChan := make(chan os.Signal, 1)
	// global stop channel. Close signal will be sent to broadvast a shutdown to everything waiting for it to stop
	stopChan := make(chan struct{}, 1)

	// Handle termination signals and shutdown gracefully
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChan
		log.Infof("Signal received: %v", sig)
		log.Infoln("Stopping autoscaler gracefully")
		close(stopChan)
	}()

	c := controller.NewController(opts, stopChan)
	c.RunForever(true)
}
