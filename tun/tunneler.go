package tun

import (
	"encoding/json"
	"io"

	"os"
)

// TunnelTargets are ec2 instances associate with elastic beanstalk that we can use to tunnel through
type TunnelTarget struct {
	// The name of the tunnel target
	Name string `json:"name"`
	// The URL to forward to
	RdsUrl string `json:"rds_url"`
	// The region to connect to
	Region string `json:"region"`
	// The name of the elastic beanstalk environment we will look up
	EbEnvName string `json:"eb_env"`
}

// Get all tunnel targets
func GetTunnelTargets() ([]TunnelTarget, error) {
	// Load from a rds_runnel.json file in the current directory
	var targets []TunnelTarget = make([]TunnelTarget, 5)

	// open the file
	file, err := os.Open("rds_tunnel.json")
	if err != nil {
		return nil, err
	}

	defer file.Close()
	// read the file

	allBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// parse the json
	if err := json.Unmarshal(allBytes, &targets); err != nil {
		return nil, err
	}

	return targets, nil

}
