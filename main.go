/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"

	"github.com/alauda/captain/pkg/chartrepo"
	"github.com/alauda/captain/pkg/util"

	"github.com/alauda/captain/pkg/webhook"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/alauda/captain/pkg/config"
	"github.com/alauda/captain/pkg/controller"
	"github.com/alauda/captain/pkg/helm"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	// wait for Makefile to write this
	version = ""
)

func main() {
	var options config.Options
	options.BindFlags()
	// options := getOptionsFromFlags()
	flag.Parse()
	config.FixKlogFlags()
	//fixKlogFlags()
	log.SetLogger(zap.Logger(true))

	// print version and exist
	if options.PrintVersion {
		fmt.Printf("Captain version: %s\n", version)
		return
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	// init helm dirs
	helm.Init()

	// init kube config
	cfg, err := clientcmd.BuildConfigFromFlags(options.MasterURL, options.KubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// create manager
	mgr, err := manager.New(cfg, options.Options)
	if err != nil {
		klog.Fatal("create controller-runtime manager error: ", err)
	}

	// add helm repo syncer
	// if err := mgr.Add(helm.NewDefaultIndexSyncer()); err != nil {
	//	klog.Fatal("add helm repo syncer error: ", err)
	// }

	// install HelmRequest CRD
	if err := util.InstallCRDIfRequired(cfg, options.InstallCRD); err != nil {
		klog.Fatalf("Error install CRD: %s", err.Error())
	}

	// install default chartrepo
	if options.InstallStableRepo {
		if err := chartrepo.InstallDefaultChartRepo(cfg, options.ChartRepoNamespace); err != nil {
			klog.Fatal("error install default helm repo:", err)
		}
		klog.Info("create default chart repo")

	}

	// create controller
	_, err = controller.NewController(mgr, &options, stopCh)
	if err != nil {
		klog.Fatalf("create controller error: %s", err.Error())
	}

	// add webhook
	if options.EnableValidateWebhook {
		if err := webhook.RegisterHandlers(mgr); err != nil {
			klog.Fatal("register handlers for webhook error : ", err)
		}
	}

	// run controller manager
	if err = mgr.Start(stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
