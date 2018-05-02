package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/bborbe/k8s-manifest-check/check"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	args := flag.Args()
	glog.V(4).Infof("found %d args to validate", len(args))
	if len(args) == 0 {
		fmt.Println("missing arg")
		os.Exit(1)
	}
	for _, arg := range args {
		glog.V(4).Infof("handle manifest %s", arg)
		if err := check.Check(arg); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
	glog.V(1).Infof("all manifest are valid")
}
