package testservicetesting

import (
	"fmt"
	"testing"

	TestClient "github.com/TAULargeScaleWorkshop/DTOY/services/test-service/client"
)

func TestTester1(t *testing.T) {
	c1 := TestClient.NewTestServiceClient()
	ret, err := c1.HelloToUser("Obayda")
	if err != nil {
		t.Errorf("Failed to say hello to user")
	}
	if ret != "Hello Obayda" {
		t.Errorf("Expected value: Hello Obayda, Got: %s", ret)
	}

	c2 := TestClient.NewTestServiceClient()
	ret, err = c2.HelloToUser("Deeb")
	if err != nil {
		t.Errorf("Failed to say hello to user")
	}

	if ret != "Hello Deeb" {
		t.Errorf("Expected value: Hello Deeb, Got: %s", ret)
	}

	ret, err = c1.HelloWorld()
	if err != nil {
		t.Errorf("Failed to say hello to world")
	}
	if ret != "Hello World" {
		t.Errorf("Expected value: Hello World, Got: %s", ret)
	}

	links, err := c2.ExtractLinksFromURL("https://www.microsoft.com", 1)
	if err != nil {
		t.Errorf("Failed to extract links from URL")
	}

	fmt.Printf("Links: %v\n", links)

	get_links, getLinksErr := c2.ExtractLinksFromURLAsync("https://www.microsoft.com", 1)
	err = c1.Store("Obayda", "Production down")
	if err != nil {
		t.Errorf("Failed to store key-value pair")
	}

	val, err := c1.Get("Obayda")
	if err != nil {
		t.Errorf("Failed to GET value %v\n", err)
	}
	if val != "Production down" {
		t.Errorf("Expected value: Production down, Got: %s", val)
	}

	if getLinksErr == nil {
		links, err = get_links()
		if err != nil {
			t.Errorf("Failed to extract links from URL %v\n", err)
		}
		fmt.Printf("Async Links: %v\n", links)
	} else {
		t.Errorf("Failed to extract links, request was not sent correctly %v\n", getLinksErr)
	}

	val, err = c2.Get("Obayda")
	if err != nil {
		t.Errorf("Failed to get value")
	}
	if val != "Production down" {
		t.Errorf("Expected value: Production down, Got: %s", val)
	}

	getRand, err := c1.WaitAndRand(3)
	if err != nil {
		t.Errorf("Failed to wait and rand")
	}

	randNum, err := getRand()
	if err != nil {
		t.Errorf("Failed to get random number")
	}

	fmt.Printf("Random number: %d\n", randNum)

	get_links, getLinksErr = c1.ExtractLinksFromURLAsync("https://apple.com", 1)

	err = c2.Store("Ahmad", "Khalila")
	if err != nil {
		t.Errorf("Failed to store key-value pair")
	}

	val, err = c1.Get("Ahmad")
	if err != nil {
		t.Errorf("Failed to get value")
	}

	if val != "Khalila" {
		t.Errorf("Expected value: Khalila, Got: %s", val)
	}

	if getLinksErr != nil {
		t.Errorf("Failed to extract links from URL")
	} else {
		links, err = get_links()
		if err != nil {
			t.Errorf("Failed to extract links from URL")
		}

		fmt.Printf("Async Links: %v\n", links)
	}
}
