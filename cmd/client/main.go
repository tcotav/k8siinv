package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"k8s.io/client-go/kubernetes"

	"github.com/spf13/viper"
	"github.com/tcotav/k8siinv/client"
	"github.com/tcotav/k8siinv/types"
)

type ClientConfig struct {
	ClusterName       string `json:"clusterName"`
	RunOutsideCluster bool   `json:"runOutsideCluster"`
	ConnectInfo       struct {
		Url   string `json:"url"`
		Retry int    `json:"retry"`
	} `json:"connectinfo"`
}

func main() {
	viper.SetConfigName("clientconfig")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/k8siinv/")
	viper.AddConfigPath("$HOME/.k8siinv")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("fatal error config file: %w", err)
	}

	var config ClientConfig
	viper.Unmarshal(&config)
	var clientset *kubernetes.Clientset
	if config.RunOutsideCluster {
		// create the clientset
		clientset, err = client.GetOutClusterConfig()
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		// this is the expected way to run this -- as a cronjob
		// create the clientset for in-cluster
		clientset, err = client.GetInClusterConfig()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	clusterImages, err := client.GetClusterInventory(clientset, config.ClusterName)
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
	client.RetryMax = config.ConnectInfo.Retry

	req, err := retryablehttp.NewRequest(http.MethodPost, config.ConnectInfo.Url, bytes.NewBuffer(b))
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
