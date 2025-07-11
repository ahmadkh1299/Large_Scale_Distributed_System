package TestServiceServant

import (
	"fmt"
	"math/rand"
	"time"

	metaffi "github.com/MetaFFI/lang-plugin-go/api"
	"github.com/MetaFFI/plugin-sdk/compiler/go/IDL"
	CacheServiceClient "github.com/TAULargeScaleWorkshop/DTOY/services/cache-service/client"
	service "github.com/TAULargeScaleWorkshop/DTOY/services/common"
	"github.com/TAULargeScaleWorkshop/DTOY/utils"
)

var pythonRuntime *metaffi.MetaFFIRuntime
var crawlerModule *metaffi.MetaFFIModule
var extract_links_from_url func(...interface{}) ([]interface{}, error)

var cacheMap map[string]string

func init() {
	cacheMap = make(map[string]string)

	// Load the Python3.11 runtime
	pythonRuntime = metaffi.NewMetaFFIRuntime("python311")
	err := pythonRuntime.LoadRuntimePlugin()
	if err != nil {
		msg := fmt.Sprintf("Failed to load runtime plugin: %v", err)
		utils.Logger.Fatalf(msg)
		panic(msg)
	}
	// Load the Crawler module
	crawlerModule, err = pythonRuntime.LoadModule("./crawler.py")
	if err != nil {
		msg := fmt.Sprintf("Failed to load ./crawler/crawler.py module: %v", err)
		utils.Logger.Fatalf(msg)
		panic(msg)
	}
	// Load the crawler function
	extract_links_from_url, err = crawlerModule.Load("callable=extract_links_from_url",
		[]IDL.MetaFFIType{IDL.STRING8, IDL.INT64}, // parameters types
		[]IDL.MetaFFIType{IDL.STRING8_ARRAY})      // return type
	if err != nil {
		msg := fmt.Sprintf("Failed to load extract_links_from_url function: %v", err)
		utils.Logger.Fatalf(msg)
		panic(msg)
	}
}

func HelloWorld() string {
	return "Hello World"
}

func HelloToUser(username string) string {
	return fmt.Sprintf("Hello %s", username)
}

func Store(key string, value string) error {
	c := CacheServiceClient.NewCacheServiceClient(service.LoadRegistryAddresses())
	err := c.Set(key, value)
	if err != nil {
		return fmt.Errorf("Failed to store key-value pair in cache: %v", err)
	}
	return nil
}

func Get(key string) (string, error) {
	c := CacheServiceClient.NewCacheServiceClient(service.LoadRegistryAddresses())
	value, err := c.Get(key)
	if err != nil {
		return "", fmt.Errorf("Failed to get value from cache: %v", err)
	}
	return value, nil
}

func WaitAndRand(seconds int32, sendToClient func(x int32) error) error {
	time.Sleep(time.Duration(seconds) * time.Second)
	return sendToClient(int32(rand.Intn(10)))
}

func ExtractLinksFromURL(url string, depth int32) ([]string, error) {
	// Call Python's extract_links_from_url.
	res, err := extract_links_from_url(url, int64(depth))
	if err != nil {
		return nil, err
	}
	return res[0].([]string), nil
}
