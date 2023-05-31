package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}

// Configuration represents the internal user configuration for OCI, including the required values for the instance.
type Configuration struct {
	// General
	LogLevel string `mapstructure:"log_level"`
	// Oracle Cloud / User Configuration
	UserId      string `mapstructure:"oci_user_id"`     // required
	TenancyId   string `mapstructure:"oci_tenancy_id"`  // required
	SubnetId    string `mapstructure:"oci_subnet_id"`   // required
	ImageId     string `mapstructure:"oci_image_id"`    // required
	Fingerprint string `mapstructure:"oci_fingerprint"` // required
	OciRegion   string `mapstructure:"oci_region"`      // required
	KeyPath     string `mapstructure:"key_path"`        // required
	// VM configuration, optional
	Shape       string `mapstructure:"shape"`
	DisplayName string `mapstructure:"display_name"`
	CPUs        int    `mapstructure:"cpus"`
	VolumeGb    int    `mapstructure:"volume_size"`
	SSHKey      string `mapstructure:"ssh_key"`
	// Creation Intervals
	CreateIntervalSeconds int `mapstructure:"create_interval_seconds"`
	ZoneIntervalSeconds   int `mapstructure:"create_zone_seconds"`

	client  core.ComputeClient `mapstructure:"-"`
	started time.Time          `mapstructure:"-"`
}

func (s *Configuration) Read() error {
	if s == nil {
		return fmt.Errorf("empty config")
	}
	s.started = time.Now()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetDefault("key_path", "oci.key")
	viper.SetDefault("shape", "VM.Standard.A1.Flex")
	viper.SetDefault("display_name", strconv.FormatInt(time.Now().UnixMilli(), 10))
	viper.SetDefault("cpus", 4)
	viper.SetDefault("volume_size", 50)
	viper.SetDefault("create_interval_seconds", 60)
	viper.SetDefault("create_zone_seconds", 10)

	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
		default:
			return err
		}
	}

	BindEnvs(*s)
	if err := viper.Unmarshal(s); err != nil {
		return err
	}

	return s.Validate()
}

func (s *Configuration) Region() (string, error) {
	return s.OciRegion, nil
}

func (s *Configuration) TenancyOCID() (string, error) {
	return s.TenancyId, nil
}

func (s *Configuration) KeyID() (string, error) {
	tenancy, err := s.TenancyOCID()
	if err != nil {
		return "", err
	}

	user, err := s.UserOCID()
	if err != nil {
		return "", err
	}

	fingerprint, err := s.KeyFingerprint()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", tenancy, user, fingerprint), nil
}

func (s *Configuration) KeyFingerprint() (string, error) {
	return s.Fingerprint, nil
}

func (s *Configuration) UserOCID() (string, error) {
	return s.UserId, nil
}

func (s *Configuration) PrivateRSAKey() (*rsa.PrivateKey, error) {
	f, err := os.ReadFile(s.KeyPath)
	if err != nil {
		return nil, err
	}

	block, rest := pem.Decode(f)
	if len(rest) > 0 {
		return nil, fmt.Errorf("Invalid pem keyfile, rest: %v", rest)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	return key.(*rsa.PrivateKey), err
}

func (s *Configuration) Validate() error {
	level, err := logrus.ParseLevel(s.LogLevel)
	if err != nil {
		return err
	}

	logrus.SetLevel(level)
	if len(s.SubnetId) <= 0 || !strings.Contains(s.SubnetId, ".subnet.") {
		return fmt.Errorf("invalid subnet id please specify with OCI_SUBNET_ID. It should look similar to ocid1.vcn.oc1.your-region.verylongrandomstring, config %v", s)
	}

	if len(s.ImageId) <= 0 || !strings.Contains(s.ImageId, ".image.") {
		return fmt.Errorf("invalid image id please specify with OCI_IMAGE_ID. It should look similar to ocid1.image.oc1.your-region.verylongrandomstring")
	}

	if len(s.UserId) <= 0 || !strings.Contains(s.UserId, ".user.") {
		return fmt.Errorf("invalid user id please specify with OCI_USER_ID. It should look similar to ocid1.user.oc1.your-region.verylongrandomstring")
	}

	if len(s.TenancyId) <= 0 || !strings.Contains(s.TenancyId, ".tenancy.") {
		return fmt.Errorf("invalid user id please specify using OCI_TENANCY_ID. It should look similar to ocid1.tenancy.oc1.your-region.verylongrandomstring")
	}

	if len(s.SSHKey) <= 0 {
		return fmt.Errorf("please specify your SSH key using SSH_KEY env. It should look similar to ssh-rsa verylongrstring user@example.com")
	}

	if len(s.OciRegion) <= 0 {
		return fmt.Errorf("please specify your region using OCI_REGION env")
	}

	if len(s.Fingerprint) <= 0 || !strings.Contains(s.Fingerprint, ":") {
		return fmt.Errorf("please specify your fingerprint matching the supplied API Key")
	}

	if !strings.Contains(s.ImageId, s.OciRegion) {
		return fmt.Errorf("OCI_IMAGE_ID must contain the region identifier")
	}

	if !strings.Contains(s.SubnetId, s.OciRegion) {
		return fmt.Errorf("OCI_SUBNET_ID must contain the region identifier")
	}

	return nil
}

// Copied from some github issue which I can't find anymore unfortunately (somewhere around spf13/viper).
// Fixes some issues with the env binding of Viper...
func BindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			tv = strings.ToLower(t.Name)
		}
		if tv == "-" {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			BindEnvs(v.Interface(), append(parts, tv)...)
		default:
			// Bash doesn't allow env variable names with a dot so
			// bind the double underscore version.
			keyDot := strings.Join(append(parts, tv), ".")
			keyUnderscore := strings.Join(append(parts, tv), "_")
			if err := viper.BindEnv(keyDot, strings.ToUpper(keyUnderscore)); err != nil {
				logrus.Errorf("Failed to bind %v to %v: %v\n", keyDot, strings.ToUpper(keyUnderscore), err)
			}
		}
	}
}
