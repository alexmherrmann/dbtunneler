package monitor

import (
	"encoding/json"

	"os"
)

type (
	RemoteSpec struct {
		// The local port to listen on
		LocalPort string `json:"local_port"`

		// The host to connect to
		RemoteHost string `json:"remote_host"`

		// The port to connect to
		RemotePort string `json:"remote_port"`
	}

	// SSM configuration
	SsmConfig struct {
		RemoteSpec
		// The name of the ec2 instance to connect to
		InstanceName string `json:"instance_name"`
	}

	EbSsmConfig struct {
		RemoteSpec
		// The name of the elastic beanstalk environment to connect to
		EnvironmentName string `json:"environment_name"`
	}

	// SSH configuration
	SshConfig struct {
		RemoteSpec
		// The username to connect with
		Username string `json:"username"`
		// The path to the private key to use
		KeyPath string `json:"key_path"`

		// TODO: Password (for Username and KeyPath) support? Some way to ask for it so it's not hardcoded?
	}
)

// TunnelTargets are ec2 instances associate with elastic beanstalk that we can use to tunnel through
type TunnelTarget struct {
	// The name of the tunnel target
	Name string `json:"name"`

	// The ssm configuration
	SsmConfig *SsmConfig `json:"ssm_config,omitempty"`

	// The eb ssm configuration
	EbSsmConfig *EbSsmConfig `json:"eb_ssm_config,omitempty"`

	// The ssh configuration
	SshConfig *SshConfig `json:"ssh_config,omitempty"`
}

// Get all tunnel targets
func GetTunnelTargets(filename string) ([]TunnelTarget, error) {
	// Load from a rds_runnel.json file in the current directory
	targets := struct {
		Targets []TunnelTarget `json:"tunnels"`
	}{
		make([]TunnelTarget, 0, 10),
	}

	// open the file
	fileContents, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// parse the json
	if err := json.Unmarshal(fileContents, &targets); err != nil {
		return nil, err
	}

	return targets.Targets, nil

}
