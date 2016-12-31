package main

// this needs a lot of work yet

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/meta"
	"k8s.io/client-go/pkg/api/v1"
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

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := clientset.CoreV1().Events("default").Watch(v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Stop()

	// handle sigterm and interrupt
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
	accessor := meta.NewAccessor()
	for message := range channel {
		kind, err := accessor.Kind(message.Object)
		if err != nil {
			log.Println(err)
		}
		namespace, err := accessor.Namespace(message.Object)
		if err != nil {
			log.Println(err)
		}
		name, err := accessor.Name(message.Object)
		if err != nil {
			log.Println(err)
		}
		log.Printf("%s, %s, %s, %s, %v\n\n", message.Type, kind, namespace, name, message.Object)
	}
}
