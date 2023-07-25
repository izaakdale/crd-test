package main

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	dynCli := dynamic.NewForConfigOrDie(config)

	bookInformer := cache.NewSharedIndexInformer(&cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			return dynCli.Resource(
				schema.GroupVersionResource{
					Group:    "example.com",
					Version:  "v1",
					Resource: "books",
				},
			).List(context.Background(), v1.ListOptions{})
		},
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return dynCli.Resource(
				schema.GroupVersionResource{
					Group:    "example.com",
					Version:  "v1",
					Resource: "books",
				},
			).Watch(context.Background(), v1.ListOptions{})
		},
	},
		&unstructured.Unstructured{},
		0,
		cache.Indexers{},
	)

	bookInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			book := obj.(*unstructured.Unstructured)

			var intObj struct {
				Spec struct {
					Author string `json:"author,omitempty"`
					Title  string `json:"title,omitempty"`
				} `json:"spec,omitempty"`
			}

			err := runtime.DefaultUnstructuredConverter.FromUnstructured(book.Object, &intObj)
			if err != nil {
				panic(err)
			}

			fmt.Printf("New book object: Author: %s, Title: %s\n", intObj.Spec.Author, intObj.Spec.Title)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Handle updates here
			fmt.Printf("Update called!\n")
		},
		DeleteFunc: func(obj interface{}) {
			// Handle deletions here
			fmt.Printf("Delete called\n")
		},
	})

	// Start the informer
	stopCh := make(chan struct{})
	defer close(stopCh)
	go bookInformer.Run(stopCh)

	// Wait for the controller to stop
	for range stopCh {
		os.Exit(0)
	}
}
