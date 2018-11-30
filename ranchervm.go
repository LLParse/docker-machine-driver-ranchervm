package ranchervm

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	api "github.com/rancher/vm/pkg/apis/ranchervm/v1alpha1"
	"github.com/rancher/vm/pkg/server"
	"github.com/rancher/vm/pkg/server/client"
)

const (
	defaultSSHUser = "docker"
)

// Driver is the RancherVM Driver struct
type Driver struct {
	*drivers.BaseDriver
	Endpoint           string
	InsecureSkipVerify bool
	AccessKey          string
	SecretKey          string
	CPU                int
	MemoryMiB          int
	Image              string
	SSHKeyName         string
	SSHKeyDelete       bool
	EnableNoVNC        bool
	NodeName           string

	LonghornBacking        bool
	LonghornVolumeSize     string
	LonghornReplicaCount   int
	LonghornReplicaTimeout int

	client *client.RancherVMClient
}

// NewDriver constructs a new RancherVM Driver
func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) getClient() *client.RancherVMClient {
	if d.client == nil {
		endpoint := strings.TrimSuffix(d.Endpoint, "/")
		d.client = client.NewRancherVMClient(endpoint, d.AccessKey, d.SecretKey, d.InsecureSkipVerify)
	}
	return d.client
}

func randomString(n int, alphabet string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[r.Intn(len(alphabet))]
	}
	return string(b)
}

func generateSSHKeyName(name string) string {
	suffix := randomString(5, "0123456789abcdef")
	return strings.Join([]string{name, suffix}, "-")
}

func generateSSHKey(path string) (string, error) {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("Desired directory for SSH keys does not exist: %s", err)
		}

		kp, err := ssh.NewKeyPair()
		if err != nil {
			return "", fmt.Errorf("Error generating key pair: %s", err)
		}
		if err := kp.WriteToFile(path, fmt.Sprintf("%s.pub", path)); err != nil {
			return "", fmt.Errorf("Error writing keys to file(s): %s", err)
		}
		return string(kp.PublicKey), nil
	}

	return "", fmt.Errorf("Key pair already exists: %s", path)
}

// Create a host using the driver's config
func (d *Driver) Create() error {

	if d.SSHKeyName == "" {
		keyName := generateSSHKeyName(d.MachineName)
		publicKey, err := generateSSHKey(d.GetSSHKeyPath())
		if err != nil {
			return err
		}
		err = d.getClient().CredentialCreate(keyName, publicKey)
		if err != nil {
			return err
		}
		d.SSHKeyName = keyName
		// If keypair name wasn't specified, we'll assume the keypair is
		// disposable and delete it when the machine is deleted
		d.SSHKeyDelete = true

		// FIXME: ranchervm creates vm pod before informer cache receives the new credential
		time.Sleep(3 * time.Second)

	} else {
		credential, err := d.getClient().CredentialGet(d.SSHKeyName)
		if err != nil {
			return err
		}
		if credential == nil {
			publicKey, err := generateSSHKey(d.GetSSHKeyPath())
			if err != nil {
				return err
			}
			err = d.getClient().CredentialCreate(d.SSHKeyName, publicKey)
			if err != nil {
				// A race exists when creating many machines concurrently with
				// a named, but not yet generated keypair. We must therefore
				// tolerate any 409 conflict errors received in this context.
				if !strings.Contains(err.Error(), http.StatusText(http.StatusConflict)) {
					return err
				}
			}
			// When generating a named keypair, do NOT automatically delete it
			// because the keypair is expected to be reused by other machines
			d.SSHKeyDelete = false

			// FIXME: ranchervm creates vm pod before informer cache receives the new credential
			time.Sleep(3 * time.Second)
		} else {
			// TODO: verify we have the private key. The public key might've
			// been uploaded manually by user, in which case we can't use it
		}
	}

	volume := api.VolumeSource{}
	if d.LonghornBacking {
		volume.Longhorn = &api.LonghornVolumeSource{
			Size:                d.LonghornVolumeSize,
			BaseImage:           d.Image,
			NumberOfReplicas:    d.LonghornReplicaCount,
			StaleReplicaTimeout: d.LonghornReplicaTimeout,
		}
	} else {
		volume.EmptyDir = &api.EmptyDirVolumeSource{}
	}

	return d.getClient().InstanceCreate(server.Instance{
		Name:        d.MachineName,
		Cpus:        d.CPU,
		Memory:      d.MemoryMiB,
		Image:       d.Image,
		Action:      string(api.ActionStart),
		PublicKeys:  []string{d.SSHKeyName},
		HostedNovnc: d.EnableNoVNC,
		NodeName:    d.NodeName,
		Volume:      volume,
	}, 1)
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "ranchervm"
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:  "ranchervm-endpoint",
			Usage: "RancherVM endpoint",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ranchervm-access-key",
			Usage: "Rancher API Access Key",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "ranchervm-secret-key",
			Usage: "Rancher API Secret Key",
			Value: "",
		},
		mcnflag.BoolFlag{
			Name:  "ranchervm-insecure-skip-verify",
			Usage: "Skip TLS certificate verification for HTTP requests to RancherVM",
		},
		mcnflag.StringFlag{
			Name:  "ranchervm-ssh-user",
			Usage: "SSH user",
			Value: "ubuntu",
		},
		mcnflag.IntFlag{
			Name:  "ranchervm-ssh-port",
			Usage: "SSH port",
			Value: drivers.DefaultSSHPort,
		},
		mcnflag.StringFlag{
			Name:  "ranchervm-ssh-key-name",
			Usage: "Use a shared SSH key",
		},
		mcnflag.StringFlag{
			Name:  "ranchervm-ssh-key-path",
			Usage: "Path to private SSH key",
		},
		mcnflag.IntFlag{
			Name:  "ranchervm-cpu-count",
			Usage: "Number of CPUs",
			Value: 1,
		},
		mcnflag.IntFlag{
			Name:  "ranchervm-memory-mib",
			Usage: "Memory in MiB",
			Value: 1024,
		},
		mcnflag.StringFlag{
			Name:  "ranchervm-image",
			Usage: "Docker image containing qcow2 disk image",
			Value: "llparse/vm-ubuntu:rancher-2.1.1",
		},
		mcnflag.BoolFlag{
			Name:  "ranchervm-novnc",
			Usage: "Enable NoVNC, a browser-based VNC client accessible from Rancher UI",
		},
		mcnflag.StringFlag{
			Name:  "ranchervm-node-name",
			Usage: "Name of Kubernetes node to schedule machine to",
		},
		mcnflag.BoolFlag{
			Name:  "ranchervm-longhorn",
			Usage: "Use Longhorn storage provider instead of host filesystem",
		},
		// TODO longhorn should eventually infer size from disk image
		mcnflag.StringFlag{
			Name:  "ranchervm-longhorn-image-size",
			Usage: "Size of the qcow2 disk image, currently required by Longhorn",
			Value: "50Gi",
		},
		mcnflag.IntFlag{
			Name:  "ranchervm-longhorn-replica-count",
			Usage: "Number of replicas to back Longhorn volume with",
			Value: 3,
		},
		mcnflag.IntFlag{
			Name:  "ranchervm-longhorn-replica-timeout",
			Usage: "Time (in seconds) to wait before replacing an unresponsive replica",
			Value: 30,
		},
	}
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	instance, err := d.getClient().InstanceGet(d.MachineName)
	if err != nil {
		return "", err
	}
	if instance.Status.IP == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	d.IPAddress = instance.Status.IP
	return instance.Status.IP, nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		log.Warnf("Failed to get IP: %s", err)
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	instance, err := d.getClient().InstanceGet(d.MachineName)
	if err != nil {
		return state.None, err
	}

	switch instance.Status.State {
	case api.StatePending:
		return state.Starting, nil
	case api.StateRunning:
		return state.Running, nil
	case api.StateStopping:
		return state.Stopping, nil
	case api.StateStopped:
		return state.Stopped, nil
	case api.StateTerminating:
		return state.Stopped, nil
	case api.StateTerminated:
		return state.Stopped, nil
	case api.StateMigrating:
		return state.Running, nil
	case api.StateError:
		return state.Error, nil
	}
	return state.None, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return d.getClient().InstanceStop(d.MachineName)
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	instance, err := d.getClient().InstanceGet(d.MachineName)
	if err != nil {
		return err
	}
	if instance != nil {
		return fmt.Errorf("MachineName %s already taken", d.MachineName)
	}
	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	if d.SSHKeyDelete {
		if err := d.getClient().CredentialDelete(d.SSHKeyName); err != nil {
			return err
		}
	}

	if err := d.getClient().InstanceDelete(d.MachineName); err != nil {
		return err
	}

	t := time.NewTicker(3 * time.Second)
	for _ = range t.C {
		if instance, err := d.getClient().InstanceGet(d.MachineName); err != nil {
			return err
		} else if instance == nil {
			break
		}
	}
	return nil
}

// ResolveStorePath returns a unique or shared store path
func (d *Driver) ResolveStorePath(file string) string {
	if d.SSHKeyName == "" {
		return filepath.Join(d.StorePath, "machines", d.MachineName, file)
	}
	return filepath.Join(d.StorePath, "machines",
		strings.Join([]string{d.DriverName(), d.SSHKeyName, file}, "."))
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Endpoint = flags.String("ranchervm-endpoint")
	d.AccessKey = flags.String("ranchervm-access-key")
	d.SecretKey = flags.String("ranchervm-secret-key")
	d.InsecureSkipVerify = flags.Bool("ranchervm-insecure-skip-verify")
	d.CPU = flags.Int("ranchervm-cpu-count")
	d.MemoryMiB = flags.Int("ranchervm-memory-mib")
	d.Image = flags.String("ranchervm-image")
	d.EnableNoVNC = flags.Bool("ranchervm-novnc")
	d.NodeName = flags.String("ranchervm-node-name")
	d.SSHKeyName = flags.String("ranchervm-ssh-key-name")
	d.SSHKeyPath = flags.String("ranchervm-ssh-key-path")
	d.SSHUser = flags.String("ranchervm-ssh-user")
	d.SSHPort = flags.Int("ranchervm-ssh-port")
	d.LonghornBacking = flags.Bool("ranchervm-longhorn")
	d.LonghornVolumeSize = flags.String("ranchervm-longhorn-image-size")
	d.LonghornReplicaCount = flags.Int("ranchervm-longhorn-replica-count")
	d.LonghornReplicaTimeout = flags.Int("ranchervm-longhorn-replica-timeout")
	return nil
}

// Start a host
func (d *Driver) Start() error {
	return d.getClient().InstanceStart(d.MachineName)
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	return d.getClient().InstanceStop(d.MachineName)
}
