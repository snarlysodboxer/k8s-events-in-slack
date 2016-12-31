package main

// this is not working at all yet

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/runtime"
	// "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = flag.String("kube-config", "/home/my-user/.kube/config", "absolute path to your kubeconfig file")
	incluster  = flag.Bool("in-cluster", false, "running in the cluster or not")
)

func getConfig() *rest.Config {
	if *incluster {
		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
		return config
	} else {
		// uses the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			log.Fatal(err)
		}
		return config
	}
}

func main() {
	flag.Parse()
	config := getConfig()

	// TODO failing here
	client, err := dynamic.NewClient(config)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	resourceClient := client.Resource(&v1.APIResource{}, "default")
	watcher, err := resourceClient.Watch(*new(runtime.Object))
	if err != nil {
		log.Fatal(err)
	}

	for {
		defer watcher.Stop()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func() {
			for sig := range ch {
				log.Printf("Stopped with signal: %s", sig)
				watcher.Stop()
				os.Exit(0)
			}
		}()

		channel := watcher.ResultChan()
		for message := range channel {
			object := message.Object
			log.Printf("%s, %v\n\n", message.Type, object)
		}
	}
}
