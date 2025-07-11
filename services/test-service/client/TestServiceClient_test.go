package TestServiceClient

import (
	"testing"
)

func TestHelloWorld(t *testing.T) {
	c := NewTestServiceClient()
	r, err := c.HelloWorld()
	if err != nil {
		t.Fatalf("could not call HelloWorld: %v", err)
		return
	}
	t.Logf("Response: %v", r)
}

func TestHelloToUser(t *testing.T) {
	c := NewTestServiceClient()
	r, err := c.HelloToUser("GreenFuzer")
	if err != nil {
		t.Fatalf("could not call HelloToUser: %v", err)
		return
	}
	t.Logf("Response: %v", r)
}

func TestStore(t *testing.T) {
	c := NewTestServiceClient()
	err := c.Store("GreenFuzer", "Epic")
	if err != nil {
		t.Fatalf("could not call Store: %v", err)
		return
	}
	t.Logf("Response: %v", "Store succeeded")
}

func TestGet(t *testing.T) {
	c := NewTestServiceClient()
	r, err := c.Get("GreenFuzer")
	if err != nil {
		t.Fatalf("could not call Get: %v", err)
		return
	}
	t.Logf("Response: %v", r)
}

func TestWaitAndRand(t *testing.T) {
	c := NewTestServiceClient()
	resPromise, err := c.WaitAndRand(3)
	if err != nil {
		t.Fatalf("Calling WaitAndRand failed: %v", err)
		return
	}
	res, err := resPromise()
	if err != nil {
		t.Fatalf("WaitAndRand failed: %v", err)
		return
	}
	t.Logf("Returned random number: %v\n", res)
}

func TestIsAlive(t *testing.T) {
	c := NewTestServiceClient()
	r, err := c.IsAlive()
	if err != nil {
		t.Fatalf("could not call IsAlive: %v", err)
		return
	}
	t.Logf("Response: %v", r)
}

func TestExtractLinksFromURL(t *testing.T) {
	c := NewTestServiceClient()
	url := "https://www.microsoft.com"
	r, err := c.ExtractLinksFromURL(url, 1)
	if err != nil {
		t.Fatalf("could not call ExtractLinksFromURL: %v", err)
		return
	}
	t.Logf("Response: %v", r)
}

/*

package TestService

import (
	"testing"

	client "github.com/TAULargeScaleWorkshop/DTOY/services/test-service/client"
)

func TestHelloWorld(t *testing.T) {
	c := client.NewTestServiceClient("localhost:50051")
	r, err := c.HelloWorld()
	if err != nil {
		t.Fatalf("could not call HelloWorld: %v", err)
		return
	}
	t.Logf("Response: %v", r)
}

*/
