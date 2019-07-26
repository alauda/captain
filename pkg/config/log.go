package config

import (
	"flag"
	"fmt"
	"os"

	"k8s.io/klog"
)

//FixKlogFlags copy flags between glog and klog
func FixKlogFlags() {
	// klog.SetOutput(os.Stdout)

	// This code sinppet copyed from klog. About glog/klog, we have a few options
	// 1. replace glog by klog in go mod --> pass kog.Infof to helm not working
	// 2. use this, it worked, but i don't known why. Fuck klog! Fuck glog!

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// Sync the glog and klog flags.
	flag.CommandLine.VisitAll(func(f1 *flag.Flag) {
		f2 := klogFlags.Lookup(f1.Name)
		if f2 != nil {
			value := f1.Value.String()
			// default is invalid for parser...
			if f1.Name != "log_backtrace_at" || value != ":0" {
				if err := f2.Value.Set(value); err != nil {
					fmt.Printf("init klog flag %s:%s error: %s\n", f1.Name, value, err.Error())
				}
			}
		}
	})

	// why i need to set this????
	klog.SetOutput(os.Stdout)
}
