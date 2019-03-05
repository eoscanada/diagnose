package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (d *Diagnose) setupK8s() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	d.cluster = clientset

	return nil
}
