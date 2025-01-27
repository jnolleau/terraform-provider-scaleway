package scaleway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	documentdb "github.com/scaleway/scaleway-sdk-go/api/documentdb/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	telemetryDocumentDBReporting       = "telemetry_reporting"
	defaultDocumentDBInstanceTimeout   = defaultRdbInstanceTimeout
	defaultWaitDocumentDBRetryInterval = defaultWaitRDBRetryInterval
)

// documentDBAPIWithRegion returns a new documentdb API and the region for a Create request
func documentDBAPIWithRegion(d *schema.ResourceData, m interface{}) (*documentdb.API, scw.Region, error) {
	meta := m.(*Meta)
	api := documentdb.NewAPI(meta.scwClient)

	region, err := extractRegion(d, meta)
	if err != nil {
		return nil, "", err
	}

	return api, region, nil
}

// documentDBAPIWithRegionalAndID returns a new documentdb API with region and ID extracted from the state
func documentDBAPIWithRegionAndID(m interface{}, regionalID string) (*documentdb.API, scw.Region, string, error) {
	meta := m.(*Meta)
	api := documentdb.NewAPI(meta.scwClient)

	region, ID, err := parseRegionalID(regionalID)
	if err != nil {
		return nil, "", "", err
	}

	return api, region, ID, nil
}

func waitForDocumentDBInstance(ctx context.Context, api *documentdb.API, region scw.Region, id string, timeout time.Duration) (*documentdb.Instance, error) {
	retryInterval := defaultWaitDocumentDBRetryInterval
	if DefaultWaitRetryInterval != nil {
		retryInterval = *DefaultWaitRetryInterval
	}

	instance, err := api.WaitForInstance(&documentdb.WaitForInstanceRequest{
		Region:        region,
		InstanceID:    id,
		RetryInterval: &retryInterval,
		Timeout:       scw.TimeDurationPtr(timeout),
	}, scw.WithContext(ctx))

	return instance, err
}

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/DatabaseName"
func resourceScalewayDocumentDBDatabaseID(region scw.Region, instanceID string, databaseName string) (resourceID string) {
	return fmt.Sprintf("%s/%s/%s", region, instanceID, databaseName)
}

// resourceScalewayDocumentDBDatabaseName extract regional instanceID and databaseName from composed ID
// returned by resourceScalewayDocumentDBDatabaseID()
func resourceScalewayDocumentDBDatabaseName(id string) (string, string, error) {
	elems := strings.Split(id, "/")
	if len(elems) != 3 {
		return "", "", fmt.Errorf("cant parse terraform database id: %s", id)
	}

	return elems[0] + "/" + elems[1], elems[2], nil
}
