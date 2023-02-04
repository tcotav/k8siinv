package client

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/tcotav/k8siinv/types"
)

func GetClusterInventory(clientset *kubernetes.Clientset, clusterName string) (types.ClusterInventory, error) {
	generatedAt := time.Now().Format(time.RFC3339)
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return types.ClusterInventory{}, err
	}
	retList := types.ClusterInventory{Version: VERSION, ClusterName: clusterName, GeneratedAt: generatedAt, ImageState: make([]types.PodImageState, 0)}
	// get all the images -- could be 1 .. N
	imageMap := make(map[string]bool)
	for _, p := range pods.Items {
		// make sure we only add one image per pod+namespace combo as it assumes common deployment
		// how to determine if we have dupes?
		// - hashmap
		// - check if duplicate key with different final field for case <a>-<b>-<c>
		// -
		// - could check if field[-1] is 5 chars long
		nameParts := strings.Split(p.Name, SEP)
		lenNP := len(nameParts)
		// check if possibly deployment
		// nginx-deployment-9456bbbf9-mfslp
		// kube-apiserver-minikube <--- not
		// <name>-...-<replicasetid>-<podid>
		mapKey := fmt.Sprintf("%s-%s", p.Namespace, strings.Join(nameParts[:lenNP-1], SEP))

		// now check if we've seen this before
		if _, ok := imageMap[mapKey]; ok {
			// we have so next pod
			continue
		} // else we haven't so process the pod
		imageMap[mapKey] = true
		imageList := make([]string, 0)
		timeStr := p.Status.StartTime.Format(time.RFC3339)
		if err != nil {
			return types.ClusterInventory{}, err
		}
		// we use the podname, p.Name, here even if it is one of multiple from a replicaSet or Deployment
		podState := types.PodImageState{Name: p.Name, Namespace: p.Namespace, StartTime: timeStr, Images: imageList}
		for _, c := range p.Spec.Containers {
			fmt.Printf("%s - %s\n", p.Name, c.Image)
			podState.Images = append(podState.Images, c.Image)
		}
		retList.ImageState = append(retList.ImageState, podState)
	}
	return retList, nil
}

func GetInClusterConfig() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	return kubernetes.NewForConfig(config)
}

func GetOutClusterConfig() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("Error build config %s", err.Error())
	}

	// create the clientset
	return kubernetes.NewForConfig(config)
}
