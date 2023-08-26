package tun_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"rdstunneler/src/tun"
	"testing"
	"time"
)

func TestStartEbSSMProxyToEnvVar(t *testing.T) {

	envValueBytes, err := ioutil.ReadFile("test_env_name.txt")
	if err != nil {
		t.Errorf("Error opening test_env_name.txt: %s", err)
		return
	}

	envValue := string(envValueBytes)

	t.Logf("AWS_EB_NAME is %s", envValue)

	ec2, err := tun.NewEc2()
	if err != nil {
		t.Errorf("Error creating ec2: %s", err)
		return
	}

	errs := ec2.RefreshAllRegions()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("Error refreshing ec2: %s", err)
		}
		return
	}

	// get the instance for the eb env
	inst, err := ec2.GetAnInstanceForBeanstalkEnv(envValue)
	if err != nil {
		t.Errorf("Error getting instance for beanstalk: %s", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	eventualErr, err := tun.StartSSMProxy(ctx,
		*inst.InstanceId,
		"8080",
		"httpbin.org",
		"80")

	if err != nil {
		t.Errorf("Error starting proxy: %s", err)
		return
	}

	success := make(chan bool)
	// Start a goroutine that keeps trying localhost:8080 until it gets a response from httpbin.org through the proxy
	go func() {
		for {
			t.Logf("Running request")
			req := http.Request{
				Method: "GET",
				URL:    &url.URL{Scheme: "http", Host: "localhost:8080"},
				Header: http.Header{
					"Host": []string{"httpbin.org"},
				},
			}
			resp, err := http.DefaultClient.Do(&req)

			if err != nil {
				t.Logf("Error making request: %s", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if resp.StatusCode > 0 {
				t.Logf("Got response: %d", resp.StatusCode)
				success <- true
				return
			}
		}
	}()

	select {
	case <-success:
		t.Logf("Got success!")
		return
	case <-ctx.Done():
		t.Errorf("Proxy timed out after 15 seconds")
		return
	case err := <-eventualErr:
		if err != nil {
			t.Errorf("Error with running proxy: %s", err)
			return
		}
	}

}
