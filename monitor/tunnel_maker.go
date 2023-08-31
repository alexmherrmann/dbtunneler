package monitor

import (
	"context"
	"fmt"
	"sync"
	"tunny/tun"
)

var _ek2 *tun.Ec2Interactor

func lazyMakeEk2() (*tun.Ec2Interactor, error) {
	if _ek2 == nil {
		var err error
		_ek2, err = tun.NewEc2()
		if err != nil {
			return nil, err
		}

		errs := _ek2.RefreshAllRegions()
		if len(errs) > 0 {
			// Collect all the error strings into one big one
			errStrCollector := ""
			for _, err := range errs {
				errStrCollector += fmt.Sprintf("%s\n", err)
			}
			return nil, fmt.Errorf("big problems: %s", errStrCollector)
		}

	}
	return _ek2, nil

}

/*
MakeTargetIntoSomething takes a tunnel target and actually starts the tunnel.
It will not return errors but will instead report them to the monitor.

The waitgroup is used to signal when the tunnels are done quitting.
It should be waited on after calling the context cancel function.
*/
func MakeTargetIntoSomething(
	topCtx context.Context,
	target TunnelTarget,
	mon MonitoringInteractor,
	waiter *sync.WaitGroup) {
	if target.EbSsmConfig != nil {
		ek2, err := lazyMakeEk2()
		if err != nil {
			mon.ReportFatalError(target.Name, "Error creating ec2 interactor: %s", err)
		}
		instanceId, err := ek2.GetAnInstanceForBeanstalkEnv(target.EbSsmConfig.EnvironmentName)
		if err != nil {
			mon.ReportFatalError(target.Name, "Error getting instance for beanstalk: %s", err)

		}
		eventual, err := tun.StartSSMProxy(
			topCtx,
			*instanceId.InstanceId,
			target.EbSsmConfig.LocalPort,
			target.EbSsmConfig.RemoteHost,
			target.EbSsmConfig.RemotePort)
		if err != nil {
			mon.ReportFatalError(target.Name, "Error starting proxy: %s", err)

		}

		go HandleStartedErrChan(mon, target, eventual, waiter)

	} else if target.SsmConfig != nil {
		eventual, err := tun.StartSSMProxy(
			topCtx,
			target.SsmConfig.InstanceName,
			target.SsmConfig.LocalPort,
			target.SsmConfig.RemoteHost,
			target.SsmConfig.RemotePort)
		if err != nil {
			mon.ReportFatalError(target.Name, "Error starting proxy: %s", err)
		}

		go HandleStartedErrChan(mon, target, eventual, waiter)

	} else if target.SshConfig != nil {
		waiter.Done()
		panic("not implemented")

	} else {
		waiter.Done()
		mon.ReportFatalError(target.Name, "Must specify ssh, ssm, or beanstalk ssm config")

	}
}

// Handle the started connection by watching for errors
func HandleStartedErrChan(
	mon MonitoringInteractor,
	targ TunnelTarget,
	errChan <-chan error,
	waiter *sync.WaitGroup) {
	defer func() {
		waiter.Done()
	}()
	err := <-errChan
	if err != nil {
		mon.ReportFatalError(targ.Name, "Proxy %s failed: %s", targ.Name, err)
	}

}
