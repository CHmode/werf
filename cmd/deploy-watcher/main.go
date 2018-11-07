package main

import (
	"fmt"
	"time"

	"github.com/flant/dapp/pkg/lock"
	"github.com/flant/dapp/pkg/ruby2go"

	"github.com/flant/kubedog/pkg/kube"
	"github.com/flant/kubedog/pkg/kubedog"
	"github.com/flant/kubedog/pkg/monitor"
)

func main() {
	var err error

	err = lock.Init()
	if err != nil {
		panic(err)
	}

	err = kube.Init()
	if err != nil {
		panic(err)
	}

	ruby2go.RunCli("deploy-watcher", func(args map[string]interface{}) (interface{}, error) {
		namespace := args["namespace"].(string)
		if namespace == "" {
			return nil, fmt.Errorf("namespace argument required!")
		}

		resourceName := args["resourceName"].(string)
		if resourceName == "" {
			return nil, fmt.Errorf("resourceName argument required!")
		}

		timeout := args["timeout"].(float64)

		switch action := args["action"]; action {
		case "watch job":
			err := kubedog.WatchJobTillDone(resourceName, namespace, kube.Kubernetes, monitor.WatchOptions{Timeout: time.Second * time.Duration(timeout)})
			if err != nil {
				return nil, fmt.Errorf("error watching job `%s` in namespace `%s`: %s", resourceName, namespace, err)
			}
		case "watch deployment":
			err := kubedog.WatchDeploymentTillReady(resourceName, namespace, kube.Kubernetes)
			if err != nil {
				return nil, fmt.Errorf("error watching deployment `%s` in namespace `%s`: %s", resourceName, namespace, err)
			}
		default:
			return nil, fmt.Errorf("unknown action \"%s\"", action)
		}

		return nil, nil
	})
}
