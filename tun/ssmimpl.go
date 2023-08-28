package tun

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func NewEc2() (*Ec2Interactor, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		return nil, err
	}

	svc := ec2.New(sess)
	return &Ec2Interactor{svc, make([]*ec2.Instance, 0, 20)}, nil
}

type Ec2Interactor struct {
	Ec2Svc *ec2.EC2

	// The ec2 instances we've looked up
	instances []*ec2.Instance
}

// refresh all ec2 instances for a region
func fetchEC2Instances(ctx context.Context, region string) ([]*ec2.Instance, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeInstancesInput{}

	ek2 := ec2.New(sess)

	result, err := ek2.DescribeInstancesWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	var instances []*ec2.Instance
	for _, reservation := range result.Reservations {
		instances = append(instances, reservation.Instances...)
	}

	return instances, nil
}

func (e *Ec2Interactor) GetAnInstanceForBeanstalkEnv(beanstalkName string) (*ec2.Instance, error) {
	for _, instance := range e.instances {
		for _, tag := range instance.Tags {
			if *tag.Key == "elasticbeanstalk:environment-name" {
				if *tag.Value == beanstalkName {
					return instance, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no instance for %s found", beanstalkName)
}

// Go try and refresh all the regions
func (e *Ec2Interactor) RefreshAllRegions() []error {
	input := &ec2.DescribeRegionsInput{}
	log.Println("Refreshing regions, getting regions first")
	regions, err := e.Ec2Svc.DescribeRegions(input)
	if err != nil {
		return []error{err}
	}

	waitgroup := &sync.WaitGroup{}
	resultChan := make(chan []*ec2.Instance, 3)
	errChan := make(chan error, 3)

	// range over the regions
	for _, reg := range regions.Regions {
		waitgroup.Add(1)
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*45)
		log.Print("Starting refresh for region ", *reg.RegionName)
		// Go start a goroutine to refresh the instances for each region
		go func(region *ec2.Region) {
			defer waitgroup.Done()
			defer cancel()
			// Refresh the instances for each region
			result, err := fetchEC2Instances(timeoutCtx, *region.RegionName)
			if err != nil {
				if err == context.DeadlineExceeded {
					errChan <- fmt.Errorf("refreshEC2Instances for region %s timed out", *region.RegionName)
				} else {
					errChan <- err
				}
				return
			}

			log.Printf("Got %d results for region %s", len(result), *region.RegionName)
			resultChan <- result
		}(reg)
	}

	// Wait for all the goroutines to finish
	go func() {
		log.Print("Waiting for all refreshes to finish")
		waitgroup.Wait()
		close(resultChan)
		close(errChan)
	}()

	// As results come in, append them to the instances list
	for result := range resultChan {
		// log.Printf("Got %d results", len(result))
		e.instances = append(e.instances, result...)
		// log.Printf("Total instances: %d", len(e.instances))
	}

	// As errors come in, append them to the errors list
	var errors []error = nil
	for err := range errChan {
		if errors == nil {
			errors = make([]error, 0)
		}
		errors = append(errors, err)
	}

	log.Printf("Got %d instances", len(e.instances))
	return errors

}

func superKill(pid int) error {
	var cmd *exec.Cmd

	stdErrBuff := &bytes.Buffer{}
	switch runtime.GOOS {
	case "windows":
		// Taskkill is used to terminate tasks by process id or image name.
		// The `/T` option is used to terminate all child processes along with the parent process, commonly known as a tree kill.
		cmd = exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T", "/F")

	case "darwin", "linux":
		// Pgrep finds processes by a parent pid, and we're passing that list to pkill to terminate them.
		// Then, we kill the parent process.
		cmd = exec.Command("/bin/sh", "-c", "pkill -TERM -P "+strconv.Itoa(pid)+"; kill -TERM "+strconv.Itoa(pid))

	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stderr = stdErrBuff

	execErr := cmd.Run()

	if execErr != nil {
		return fmt.Errorf("error killing process (stderr below): %s\n%s", execErr, stdErrBuff.String())
	}

	return nil
}

/*
 * Will start the proxy on the given instance with the desired local port
 */
func StartSSMProxy(
	ctx context.Context,
	instanceid string,
	localport string,
	remote_host string,
	remote_port string,
) (ocmdErr <-chan error, startErr error) {

	// documentStr := fmt.Sprintf(
	// 	`'{"host":"%s","portNumber":["%s"],"localPortNumber":["%s"]}'`,
	// 	remote_host,
	// 	remote_port,
	// 	localport,
	// )

	documentStr := fmt.Sprintf(
		"host=%s,portNumber=%s,localPortNumber=%s",
		remote_host,
		remote_port,
		localport,
	)

	// Set up the command
	cmd := exec.CommandContext(
		ctx,
		"aws",
		"ssm",
		"start-session",
		"--target",
		instanceid,
		"--document-name",
		"AWS-StartPortForwardingSessionToRemoteHost",
		"--parameters",
		documentStr,
	)

	doKill := make(chan bool)

	go func(ctx context.Context) {
		var kerr error
		select {
		case <-ctx.Done():
			log.Printf("Context done, killing process")
			kerr = superKill(cmd.Process.Pid)
		case <-doKill:
			log.Printf("Killing process")
			kerr = superKill(cmd.Process.Pid)
		}
		if kerr != nil {
			log.Printf("> Error killing process: %s", kerr)
		}
	}(ctx)

	errBits := &bytes.Buffer{}
	cmd.Stderr = errBits

	cmdErr := make(chan error)
	ocmdErr = cmdErr
	// Start the command
	if err := cmd.Start(); err != nil {
		startErr = err
	}

	go func() {
		// defer func() {
		// 	kerr := superKill(cmd.Process.Pid)
		// 	if kerr != nil {
		// 		log.Printf("Error killing process after waiting: %s", kerr)
		// 	}
		// }()
		if err := cmd.Wait(); err != nil {
			// Capture the stderr and wrap it along with the original error
			cmdErr <- fmt.Errorf("%v: %s", err, errBits.String())
		}
		doKill <- true
		close(cmdErr)
	}()

	return
}
