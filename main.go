package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"github.com/mingfang/alluxio-csi/alluxio"
	"github.com/spf13/cobra"
)

var (
	endpoint string
	nodeID   string
)

func init() {
	flag.Set("logtostderr", "true")
}

func main() {

	flag.CommandLine.Parse([]string{})

	cmd := &cobra.Command{
		Use:   "Alluxio",
		Short: "CSI based Alluxio driver",
		Run: func(cmd *cobra.Command, args []string) {
			handle()
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(&nodeID, "nodeid", "", "node id")
	cmd.MarkPersistentFlagRequired("nodeid")

	cmd.PersistentFlags().StringVar(&endpoint, "endpoint", "", "CSI endpoint")
	cmd.MarkPersistentFlagRequired("endpoint")

	cmd.ParseFlags(os.Args[1:])
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func handle() {
	startReaper()

	d := alluxio.NewDriver(nodeID, endpoint)
	d.Run()
}

/*
Based on https://github.com/openshift/origin/blob/master/pkg/util/proc/reaper.go
The alluxio-fuse script nohup the Alluxio java client for the FUSE mount and then exits.
That causes the java process to become defunct after un-mounting.
 */
func startReaper() {
	glog.V(4).Infof("Launching reaper")
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGCHLD)
		for {
			// Wait for a child to terminate
			sig := <-sigs
			glog.V(4).Infof("Signal received: %v", sig)
			for {
				// Reap processes
				cpid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
				if cpid < 1 {
					break
				}

				glog.V(4).Infof("Reaped process with pid %d", cpid)
			}
		}
	}()
}