/*

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
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/alauda/captain/pkg/chartrepo"
	"github.com/alauda/captain/pkg/helm"
	"github.com/alauda/captain/pkg/util"

	"github.com/alauda/captain/controllers"
	"github.com/alauda/captain/pkg/cluster"
	"github.com/alauda/captain/pkg/config"
	"github.com/alauda/captain/pkg/controller"
	"github.com/alauda/captain/pkg/webhook"
	alaudaiov1alpha1 "github.com/alauda/helm-crds/pkg/apis/app/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = alaudaiov1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var options config.Options
	options.BindFlags()

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	// this avoid slow list in controller.... does not know why, but it works.
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	rp := time.Second * 120
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
		SyncPeriod:         &rp,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ChartRepoReconciler{
		Client:    cl,
		Log:       ctrl.Log.WithName("controllers").WithName("ChartRepo"),
		Scheme:    mgr.GetScheme(),
		Namespace: options.ChartRepoNamespace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ChartRepo")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	// legacy code....

	// init helm dirs
	helm.Init()

	// add cluster refresher
	cr := cluster.NewClusterRefresher(options.ClusterNamespace, mgr.GetConfig())
	if err := mgr.Add(cr); err != nil {
		setupLog.Error(err, "add cluster refresher runner error")
		os.Exit(1)
	}

	// install HelmRequest CRD
	if err := util.InstallCRDIfRequired(mgr.GetConfig(), options.InstallCRD); err != nil {
		setupLog.Error(err, "Error install CRD")
		os.Exit(1)
	}

	// install default chartrepo
	if options.InstallStableRepo {
		if err := chartrepo.InstallDefaultChartRepo(mgr.GetConfig(), options.ChartRepoNamespace); err != nil {
			setupLog.Error(err, "error install default chart repo")
			os.Exit(1)
		}
		setupLog.Info("create default chart repo")
	}

	// create controller
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := ctrl.SetupSignalHandler()
	_, err = controller.NewController(mgr, &options, stopCh)
	if err != nil {
		setupLog.Error(err, "create controller error")
		os.Exit(1)
	}

	// add webhook
	if options.EnableValidateWebhook {
		if err := webhook.RegisterHandlers(mgr); err != nil {
			setupLog.Error(err, "register handlers for webhook error")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
