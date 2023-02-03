package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/tcotav/imginv/types"
)

const VERSION string = "1.0.0"
const SEP string = "-"

func getClusterInventory(clientset *kubernetes.Clientset, clusterName string) (types.ClusterInventory, error) {
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

func main() {
	clusterName := "testcluster"
	// get connected to the k8s cluster
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
		log.Fatal(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	clusterImages, err := getClusterInventory(clientset, clusterName)
	if err != nil {
		log.Fatal(err.Error())
	}

	// convert it to json
	b, err := json.Marshal(clusterImages)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("%s\n", string(b))

	// set up our http client
	client := retryablehttp.NewClient()
	client.RetryMax = 10
	url := "http://localhost:8080/v1/images"

	req, err := retryablehttp.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// check return code
	if resp.StatusCode != 201 {
		log.Fatal("Request Problem, ", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	httpRet := types.HttpReturn{}
	json.Unmarshal(body, &httpRet)
	if httpRet.Message != "OK" {
		log.Fatal("Server returned error: ", httpRet.Message)
	}
	log.Println("Success")
}
