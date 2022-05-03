package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getKeys(m map[string]*schema.Resource) []string {
	j := 0
	keys := make([]string, len(m))
	for k := range m {
		keys[j] = k
		j++
	}
	return keys
}

var testProviders = New()

func TestProviderIsValid(t *testing.T) {
	if err := testProviders.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderSchemaHasNoResources(t *testing.T) {
	resources := testProviders.Resources()
	if len(resources) != 0 {
		t.Fatalf("Expected no resources but found %d", len(resources))
	}
}

func TestProviderSchemaHasNexusRawDataSource(t *testing.T) {
	dataSources := testProviders.DataSources()
	if len(dataSources) != 1 {
		t.Fatalf("Expected 1 data source but found %d", len(dataSources))
	}
	if _, ok := testProviders.DataSourcesMap["dependencies_nexus_raw"]; !ok {
		t.Fatal("Unable to find dependencies_nexus_raw data source in map")
	}
}
