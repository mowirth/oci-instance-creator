package main

import (
	"context"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/core"
)

func main() {
	// Config
	ctx := context.TODO()
	config := &Configuration{}
	if err := config.Read(); err != nil {
		panic(err)
	}

	client, err := core.NewComputeClientWithConfigurationProvider(config)
	if err != nil {
		panic(err)
	}

	config.client = client

	domains, err := config.ListDomains(ctx)
	if err != nil {
		panic(err)
	}

	t := time.NewTicker(time.Duration(config.CreateIntervalSeconds) * time.Second)
	fmt.Printf("Starting instance generation every %v seconds", config.CreateIntervalSeconds)
	config.createInstancesInAvailabilityZone(ctx, domains)
	for {
		select {
		case <-t.C:
			config.createInstancesInAvailabilityZone(ctx, domains)
		}
	}
}
