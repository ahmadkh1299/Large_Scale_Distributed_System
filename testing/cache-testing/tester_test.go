package cachetesting

import (
	"testing"

	CacheClient "github.com/TAULargeScaleWorkshop/DTOY/services/cache-service/client"
)

func TestCache1(t *testing.T) {
	c := CacheClient.NewCacheServiceClient(nil)
	err := c.Set("Obayda", "Production down")
	if err != nil {
		t.Errorf("Failed to set key-value pair")
	}
	val, err := c.Get("Obayda")
	if err != nil {
		t.Errorf("Failed to get value")
	}
	if val != "Production down" {
		t.Errorf("Expected value: Production down, Got: %s", val)
	}

	err = c.Set("Obayda", "Production up")
	if err != nil {
		t.Errorf("Failed to set key-value pair")
	}
	val, err = c.Get("Obayda")
	if err != nil {
		t.Errorf("Failed to get value")
	}
	if val != "Production up" {
		t.Errorf("Expected value: Production up, Got: %s", val)
	}

	err = c.Delete("Obayda")
	if err != nil {
		t.Errorf("Failed to delete key-value pair")
	}
	val, err = c.Get("Obayda")
	if err != nil {
		t.Errorf("Failed to get value")
	}
	if val != "" {
		t.Errorf("Expected value: '', Got: %s", val)
	}

	err = c.Set("Deeb", "Tibi")
	if err != nil {
		t.Errorf("Failed to set key-value pair")
	}

	err = c.Set("Ahmad", "Khalila")
	if err != nil {
		t.Errorf("Failed to set key-value pair")
	}

	err = c.Set("Obayda", "Haj")
	if err != nil {
		t.Errorf("Failed to set key-value pair")
	}

	val, err = c.Get("Deeb")
	if err != nil {
		t.Errorf("Failed to get value")
	}

	if val != "Tibi" {
		t.Errorf("Expected value: Tibi, Got: %s", val)
	}

	val, err = c.Get("Ahmad")
	if err != nil {
		t.Errorf("Failed to get value")
	}

	if val != "Khalila" {
		t.Errorf("Expected value: Khalila, Got: %s", val)
	}

	val, err = c.Get("Obayda")
	if err != nil {
		t.Errorf("Failed to get value")
	}

	if val != "Haj" {
		t.Errorf("Expected value: Haj, Got: %s", val)
	}

	err = c.Delete("Deeb")
	if err != nil {
		t.Errorf("Failed to delete key-value pair")
	}

	val, err = c.Get("Deeb")
	if err != nil {
		t.Errorf("Failed to get value")
	}

	if val != "" {
		t.Errorf("Expected value: '', Got: %s", val)
	}
}
