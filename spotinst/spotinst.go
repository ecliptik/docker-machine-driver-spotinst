package spotinst

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup"
	"github.com/spotinst/spotinst-sdk-go/service/elastigroup/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/credentials"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

const (
	driverName     = "spotinst"
	defaultSSHUser = "ubuntu"
	dockerPort     = 2376
	sshPorts       = 22
	tag            = "[SPOTINST DRIVER] "
)

type Config struct {
	Token   string
	Account string
}

type Driver struct {
	*drivers.BaseDriver
	Id                    string
	clientFactory         func() Client
	SpotinstAccount       string
	SpotinstToken         string
	SpotinstElastiGroupID string
	SSHUser               string
	PublicDNS             *string
	PrivateIpAddress      *string
	PublicIpAddress       *string
	UsePublicIPOnly       bool
	InstanceId            *string
	SpotInstanceRequest   string
}

type Client struct {
	elastigroup elastigroup.Service
}

func NewDriver(hostName, storePath string) *Driver {
	id := generateId()
	driver := &Driver{
		Id: id,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}

	driver.clientFactory = driver.BuildClient

	return driver
}

func (d *Driver) getClient() Client {
	return d.clientFactory()
}

// Validate returns an error in case of invalid configuration.
func (c *Config) Validate() error {
	return nil
}

// Client returns a new client for accessing Spotinst.
func (d *Driver) BuildClient() Client {
	config := spotinst.DefaultConfig()
	config.WithUserAgent("DockerMachine")

	var static *credentials.StaticProvider
	if d.SpotinstToken != "" || d.SpotinstAccount != "" {
		static = &credentials.StaticProvider{
			Value: credentials.Value{
				Token:   d.SpotinstToken,
				Account: d.SpotinstAccount,
			},
		}

	}
	creds := credentials.NewCredentials(static)

	if _, err := creds.Get(); err != nil {
		stdLog(ERROR, "Failed to instantiate Spotinst client: %v", err)
	}
	config.WithCredentials(creds)

	// Create a new session.
	sess := session.New(config)

	// Create a new client.
	client := &Client{
		elastigroup: elastigroup.New(sess),
	}

	return *client
}

func generateId() string {
	rb := make([]byte, 10)
	_, err := rand.Read(rb)
	if err != nil {
		log.Warnf(tag+"Unable to generate id: %s", err)
	}

	h := md5.New()
	io.WriteString(h, string(rb))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{

		mcnflag.StringFlag{
			Name:   "spotinst-token",
			Usage:  "spotinst token",
			EnvVar: "SPOTINST_TOKEN",
		},
		mcnflag.StringFlag{
			Name:   "spotinst-account",
			Usage:  "spotinst account",
			EnvVar: "SPOTINST_ACCOUNT",
		},
		mcnflag.StringFlag{
			Name:   "spotinst-elastigroup-id",
			Usage:  "spotinst elastigroup id",
			EnvVar: "SPOTINST_ELSTIGROUP_ID",
		},
		mcnflag.StringFlag{
			Name:   "spotinst-sshkey-path",
			Usage:  "spotinst sshkey path",
			EnvVar: "SPOTINST_SSHKEY_PATH",
		},
		mcnflag.BoolFlag{
			Name:   "use-public-ip",
			Usage:  "use public ip",
			EnvVar: "USE_PUBLIC_IP",
		},
		mcnflag.StringFlag{
			Name:   "ssh-user",
			Usage:  "use ssh user",
			EnvVar: "SSH_USER",
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.SpotinstAccount = flags.String("spotinst-account")
	d.SpotinstToken = flags.String("spotinst-token")
	d.SpotinstElastiGroupID = flags.String("spotinst-elastigroup-id")
	d.UsePublicIPOnly = flags.Bool("use-public-ip")
	d.SSHUser = flags.String("ssh-user")
	d.SSHKeyPath = flags.String("spotinst-sshkey-path")

	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) PrezCreateCheck() error {

	if d.SpotinstToken == "" || d.SpotinstAccount == "" {
		err := errors.New(tag + "Spotinst credentials was not provided")
		return err
	}

	if d.SSHKeyPath == "" {
		err := errors.New(tag + "Server SSH Key not provided")
		return err
	}
	return nil
}

func (d *Driver) Create() error {

	if err := d.PreCreateCheck(); err != nil {
		return err
	}

	if err := d.innerCreate(); err != nil {
		d.Remove()
		return err
	}

	return nil
}

func (d *Driver) innerCreate() error {
	stdLog(DEBUG, "creating new server for you...")
	var scaleType = "up"
	var adjustment = 1
	input := new(aws.ScaleGroupInput)
	input.Adjustment = &adjustment
	input.GroupID = &d.SpotinstElastiGroupID
	input.ScaleType = &scaleType
	output, e := d.getClient().elastigroup.CloudProviderAWS().Scale(context.Background(), input)
	if e != nil {
		stdLog(ERROR, "Client initialized failed %v", e.Error())
		return e
	}

	if len(output.Items) == 0 {
		err := errors.New("No server created as result of scale")
		log.Error(tag+"%v", err.Error())
		return err
	}

	scaleResultItem := output.Items[0]

	// Handle spotrequest
	if scaleResultItem.NewSpotRequests != nil && scaleResultItem.NewInstances == nil {
		spotInstanceRequestID := scaleResultItem.NewSpotRequests[0].SpotInstanceRequestID
		stdLog(DEBUG, "Spotrequest: %v", spotinst.StringValue(spotInstanceRequestID))

		if spotInstanceRequestID != nil {
			err := d.waitForInstanceSpot(spotInstanceRequestID)
			if err != nil {
				log.Errorf(tag + "Failed to get server from spot request " + e.Error())
				return err
			}
		} else {
			stdLog(ERROR, "Failed to get spot request")
			err := errors.New("Failed to get spot request")
			return err
		}

		errInst := d.waitForInstanceStart()

		if errInst != nil {
			stdLog(ERROR, "Instance failed to start: %v, Initiate kill", e.Error())
			return errInst
		}

	} else if scaleResultItem.NewInstances != nil {
		d.InstanceId = scaleResultItem.NewInstances[0].InstanceID
	}

	if scaleResultItem.NewInstances != nil {
		err := d.waitForInstanceStart()

		if err != nil {
			stdLog(ERROR, "Instance failed to start: %v, Initiate kill", e.Error())
			return err
		}
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {

	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(dockerPort))), nil
}

func (d *Driver) GetIP() (string, error) {

	if d.UsePublicIPOnly {
		if d.PublicIpAddress == nil {
			return "", fmt.Errorf("No public IP for instance %v", d.InstanceId)
		}
		return *d.PublicIpAddress, nil
	}

	if d.PrivateIpAddress == nil {
		return "", fmt.Errorf("No IP for instance %v", d.InstanceId)
	}

	return *d.PrivateIpAddress, nil
}

func (d *Driver) GetState() (state.State, error) {

	instance, err := d.getInstanceStatus()
	if err != nil {
		return state.Error, nil
	}
	switch *instance.Status {
	case InstanceStateNamePending:
		return state.Starting, nil
	case InstanceStateNameRunning:
		return state.Running, nil
	case InstanceStateNameStopping:
		return state.Stopping, nil
	case InstanceStateNameShuttingDown:
		return state.Stopping, nil
	case InstanceStateNameStopped:
		return state.Stopped, nil
	case InstanceStateNameTerminated:
		return state.Error, nil
	case InstanceStateNamePendingEvaluation:
		return state.Starting, nil
	case InstanceStateNameFullfiled:
		return state.Running, nil
	default:
		log.Warnf(tag+"unrecognized instance state: %v", *instance.Status)
		return state.Error, nil
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}

	return d.SSHUser
}

func (d *Driver) GetSSHPort() (int, error) {
	d.SSHPort = sshPorts
	stdLog(DEBUG, "Found SSH Port %v", d.SSHPort)
	return d.SSHPort, nil
}

func (d *Driver) GetSSHKeyPath() string {
	stdLog(DEBUG, "Found Keypath %v", d.SSHKeyPath)
	return d.SSHKeyPath
}

func (d *Driver) Start() error {
	fmt.Errorf(tag + "Spotinst not support start instance function")
	return nil
}

func (d *Driver) Stop() error {
	fmt.Errorf(tag + "Spotinst not support stop instance function")
	return nil
}

func (d *Driver) Restart() error {
	fmt.Errorf(tag + "Spotinst not support restart instance function")
	return nil
}

func (d *Driver) Kill() error {
	input := new(aws.DetachGroupInput)

	input.GroupID = &d.SpotinstElastiGroupID
	if d.InstanceId == nil {
		return nil
	}

	input.InstanceIDs = []string{*d.InstanceId}
	decrement := true
	input.ShouldDecrementTargetCapacity = &decrement
	d.getClient().elastigroup.CloudProviderAWS().Detach(context.Background(), input)

	return nil
}

func (d *Driver) Remove() error {
	return d.Kill()
}

//region Helpers
func (d *Driver) getSpotRequestStatus(spotRequestId *string) (*aws.Instance, error) {

	spotReqParam := spotinst.StringValue(spotRequestId)
	input := new(aws.StatusGroupInput)
	input.GroupID = &d.SpotinstElastiGroupID

	output, e := d.getClient().elastigroup.CloudProviderAWS().Status(context.Background(), input)

	if e != nil {
		return nil, e
	}

	if output != nil && output.Instances != nil && len(output.Instances) > 0 {
		for _, v := range output.Instances {

			spotReq := spotinst.StringValue(v.SpotRequestID)

			if spotReq == spotReqParam {

				if v != nil && v.ID != nil {
					return v, nil
				} else {
					return nil, nil
				}
			}
		}
	}

	stdLog(DEBUG, "did not find status for spot request %v", spotRequestId)
	err := errors.New("Spot Request canceled")
	return nil, err
}

func (d *Driver) getInstanceStatus() (*aws.Instance, error) {
	input := new(aws.StatusGroupInput)
	input.GroupID = &d.SpotinstElastiGroupID
	output, e := d.getClient().elastigroup.CloudProviderAWS().Status(context.Background(), input)

	if e != nil {
		return nil, e
	}

	if len(output.Instances) > 0 {
		for _, v := range output.Instances {
			if *v.ID == *d.InstanceId {
				return v, nil
			}
		}
	}

	err := errors.New("Cannot find instance " + *d.InstanceId)
	return nil, err
}

func (d *Driver) waitForInstanceStart() error {
	laps := 15
	stdLog(DEBUG, "waiting for instance Ip...", nil)
	for d.PublicIpAddress == nil && d.PrivateIpAddress == nil && laps != 0 {
		inst, e := d.getInstanceStatus()

		if e != nil {
			return e
		}

		if d.UsePublicIPOnly {
			publicIP := inst.PublicIP
			if publicIP != nil {
				stdLog(DEBUG, "Found public IP %v", publicIP)
				d.PublicIpAddress = publicIP
				return nil
			}
		} else {
			privateIP := inst.PrivateIP
			if privateIP != nil {
				stdLog(DEBUG, "Found private IP %v", privateIP)
				d.PrivateIpAddress = privateIP
				return nil
			}
		}

		laps = laps - 1
		stdLog(DEBUG, "Waiting for instance IP %v retries left", strconv.Itoa(laps))
		time.Sleep(20 * time.Second)
	}

	err := errors.New("Wait to instance " + *d.InstanceId + "reached timeout")
	return err

}

func (d *Driver) waitForInstanceSpot(spotInstanceRequestID *string) error {
	laps := 10
	stdLog(DEBUG, "waiting for spot request to get instance.. ")
	for d.InstanceId == nil && laps != 0 {
		stdLog(INFO, "Check spot request status", nil)
		instance, err := d.getSpotRequestStatus(spotInstanceRequestID)

		if err != nil {
			return err
		}

		if instance != nil {
			stdLog(DEBUG, "Instance found %v", spotinst.StringValue(instance.ID))
			d.InstanceId = instance.ID
			return nil
		}

		laps = laps - 1
		stdLog(DEBUG, "Waiting for instance  %v retries left", strconv.Itoa(laps))
		time.Sleep(20 * time.Second)

	}

	err := errors.New("Wait to spot" + *spotInstanceRequestID + "reached timeout")

	return err
}

func stdLog(logSeverity string, fmtString string, args ...interface{}) {
	switch logSeverity {
	case DEBUG:
		if args != nil {
			log.Debugf(tag+fmtString, args)
		} else {
			log.Debugf(tag + fmtString)
		}
	case INFO:
		if args != nil {
			log.Infof(tag+fmtString, args)
		} else {
			log.Infof(tag + fmtString)
		}
	case WARN:
		if args != nil {
			log.Warnf(tag+fmtString, args)
		} else {
			log.Warnf(tag + fmtString)
		}
	case ERROR:
		if args != nil {
			log.Error(tag+fmtString, args)
		} else {
			log.Error(tag + fmtString)
		}
	default:
		if args != nil {
			log.Infof(tag+fmtString, args)
		} else {
			log.Infof(tag + fmtString)
		}
	}
}

//endregion
