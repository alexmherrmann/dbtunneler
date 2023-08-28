package tun_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
	"tunny/tun"

	"golang.org/x/crypto/ssh"
)

func TestStartSshProxy(t *testing.T) {

	var sshConfig struct {
		Host     string
		Username string
		KeyPath  string
	}

	// Read the config
	readFile, err := os.ReadFile("test_ssh_config.json")
	if err != nil {
		t.Errorf("Error reading ssh config: %s", err)
		return
	}

	// Unmarshal it
	err = json.Unmarshal(readFile, &sshConfig)
	if err != nil {
		t.Errorf("Error unmarshalling test ssh config: %s", err)
		return
	}

	keyFile, err := os.ReadFile(sshConfig.KeyPath)
	if err != nil {
		t.Errorf("Error reading key file: %s", err)
		return
	}
	parsedKey, err := ssh.ParsePrivateKey(keyFile)
	if err != nil {
		t.Errorf("Error parsing key file: %s", err)
		return
	}

	// Set up the ssh realSshConfig
	realSshConfig := &ssh.ClientConfig{
		User:            sshConfig.Username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(parsedKey)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)

	defer cancel()

	t.Logf("Starting ssh tunnel")
	// Connect to the ssh server
	eventualErr, err := tun.StartSSHTunnel(
		ctx,
		realSshConfig,
		sshConfig.Host,
		"8082",
		"httpbin.org",
		"80")

	if err != nil {
		t.Errorf("Error starting ssh tunnel: %s", err)
		return
	} else {
		t.Logf("Started ssh tunnel!")
	}
	success := make(chan bool)

	// Start a goroutine to the local port, trying to make a request to httpbin.org
	go func() {
		for {
			t.Logf("Running request")
			req := http.Request{
				Method: "GET",
				URL:    &url.URL{Scheme: "http", Host: "localhost:8082"},
			}

			resp, err := http.DefaultClient.Do(&req)
			if err != nil {
				t.Logf("Error making request: %s", err)
				continue
			}
			if resp.StatusCode != 200 {
				t.Logf("Got non-200 response: %s", resp.Status)
				time.Sleep(1 * time.Second)
				continue
			}

			t.Logf("Response: %s", resp.Status)
			success <- true
			return
		}
	}()

	// Now we wait on all our channels to see what happens
	select {
	case <-success:
		t.Logf("Got success!")
		return
	case err := <-eventualErr:
		if err != nil {
			t.Errorf("Error with running proxy: %s", err)
			return
		}
	case <-ctx.Done():
		t.Errorf("Context done: %s", ctx.Err())
		return

	}

}
