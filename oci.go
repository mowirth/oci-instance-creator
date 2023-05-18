package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/identity"
)

func (s *Configuration) createInstancesInAvailabilityZone(ctx context.Context, domains identity.ListAvailabilityDomainsResponse) {
	for _, domain := range domains.Items {
		fmt.Printf("Attempting to create new instance in domain %v\n", *domain.Name)
		s.CreateInstance(ctx, domain)
		time.Sleep(time.Duration(s.ZoneIntervalSeconds) * time.Second)
	}
}

func (s *Configuration) CreateInstance(ctx context.Context, domain identity.AvailabilityDomain) {
	trueVar := true
	ocpuVar := float32(s.CPUs)
	volumeVar := int64(s.VolumeGb)

	req := core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			AvailabilityDomain: domain.Name,
			CompartmentId:      &s.TenancyId,
			Shape:              &s.Shape,
			CreateVnicDetails: &core.CreateVnicDetails{
				AssignPublicIp: &trueVar,
				SubnetId:       &s.SubnetId,
			},
			DisplayName: &s.DisplayName,
			Metadata: map[string]string{
				"ssh_authorized_keys": s.SSHKey,
			},
			ShapeConfig: &core.LaunchInstanceShapeConfigDetails{
				Ocpus: &(ocpuVar),
			},
			SourceDetails: &core.InstanceSourceViaImageDetails{
				ImageId:             &s.ImageId,
				BootVolumeSizeInGBs: &volumeVar,
			},
			IsPvEncryptionInTransitEnabled: &trueVar,
		},
	}

	resp, err := s.client.LaunchInstance(ctx, req)
	if err == nil {
		fmt.Printf("Generated instance in availability zone %v, took %v", domain.Id, time.Since(s.started).String())
		os.Exit(0)
	}

	if !strings.Contains(err.Error(), "Out of host capacity") {
		fmt.Printf("Received error from api: %v; resp: %v, \n req: %v\n\n", err, resp, req)
	}

	if strings.Contains(err.Error(), "error:TooManyRequests") {
		fmt.Printf("Received too many requests, increasing request time between instances")
		s.ZoneIntervalSeconds += 5
	}
}

func (s *Configuration) ListDomains(ctx context.Context) (identity.ListAvailabilityDomainsResponse, error) {
	domainClient, err := identity.NewIdentityClientWithConfigurationProvider(s)
	if err != nil {
		panic(err)
	}

	return domainClient.ListAvailabilityDomains(ctx, identity.ListAvailabilityDomainsRequest{
		CompartmentId:   &s.TenancyId,
		RequestMetadata: common.RequestMetadata{},
	})
}