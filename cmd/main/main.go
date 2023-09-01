package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
	"tunny/monitor"
	"tunny/webmonitor"
)

func getStdLogger() (w io.Writer, err error) {
	w, err = os.OpenFile("tunny.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	return
}

func main() {

	filename := flag.String("filename", "tunny.json", "The filename to load tunnel targets from")

	flag.Parse()

	oLog, err := getStdLogger()
	defer oLog.(io.Closer).Close()
	if err != nil {
		fmt.Printf("Error getting std logger: %s", err)
		os.Exit(1)
	}

	log.SetOutput(oLog)

	targets, err := monitor.GetTunnelTargets(*filename)
	exitCode := 0

	topCtx, cancelFunc := context.WithCancel(context.Background())

	// Handle the 'q' quit command
	go func() {
		read, err := bufio.NewReader(os.Stdin).ReadBytes('\n')
		if err != nil {
			fmt.Printf("Error reading from stdin: %s", err)
			exitCode = 1
			cancelFunc()
			return
		} else if read[0] == 'q' {
			fmt.Printf("Received quit command, shutting down\n")
			cancelFunc()
		}
	}()

	defer func() {
		cancelFunc()
		// Give enough time to hopefully let the process exit, a hack for sure
		time.Sleep(350 * time.Millisecond)
		os.Exit(exitCode)
	}()

	if err != nil {
		fmt.Printf("Error getting tunnel targets: %s", err)
		exitCode = 1

		return
	}

	waiter := &sync.WaitGroup{}

	// var mon monitor.MonitoringInteractor = &monitor.CliMonitor{}
	mon, err := webmonitor.NewWebMonitor(topCtx, targets)
	if err != nil {
		fmt.Printf("Error creating web monitor: %s", err)
		exitCode = 1
		return
	}

	for _, target := range targets {
		waiter.Add(1)
		monitor.MakeTargetIntoSomething(topCtx, target, mon, waiter)

	}

	mon.ReportGeneralMessage("All proxies started")

	// Wait for all the goroutines to finish
	waiter.Wait()

}
