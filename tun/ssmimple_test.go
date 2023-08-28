package tun_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
	"tunny/tun"
)

func TestStartWithBadParamsAndCleaned(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan, err := tun.StartSSMProxy(ctx,
		"i-12345678901234567",
		"8081",
		"httpbin.org",
		"80")

	if err != nil {
		t.Errorf("Error starting proxy: %s", err)
		return
	} else {
		t.Logf("Started proxy")
	}

	select {
	case <-ctx.Done():
		t.Errorf("Context timed out")
		return
	case err := <-errChan:
		if err != nil {
			t.Logf("Got expected error: %s", err)
			return
		}
	}

}

func TestStartEbSSMProxyToEnvVar(t *testing.T) {

	envValueBytes, err := os.ReadFile("test_env_name.txt")
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

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	eventualErr, err := tun.StartSSMProxy(ctx,
		*inst.InstanceId,
		"8080",
		"httpbin.org",
		"80")

	if err != nil {
		t.Errorf("Error starting proxy: %s", err)
		return
	} else {
		t.Logf("Started proxy")
	}

	success := make(chan bool)

	// Start a goroutine that keeps trying localhost:8080 until it gets a response from httpbin.org through the proxy
	go func() {
		for {
			t.Logf("Running request %s", time.Now().String())
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

			if resp.StatusCode == 200 {
				t.Logf("Got response: %d", resp.StatusCode)
				success <- true
				break
			} else {
				t.Logf("Got non-200 response: %d", resp.StatusCode)
				time.Sleep(1 * time.Second)
				continue
			}
		}
	}()

	select {
	case <-success:
		t.Logf("Got success!")
		return
	case <-ctx.Done():
		t.Errorf("Context timed out")
		return
	case err := <-eventualErr:
		if err != nil {
			t.Errorf("Error with running proxy: %s", err)
			return
		}
	}

}
