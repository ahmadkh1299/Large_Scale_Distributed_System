package CacheServiceServant

import (
	"fmt"
	"os"
	"strings"

	ConfigBase "github.com/TAULargeScaleWorkshop/DTOY/config"
	common "github.com/TAULargeScaleWorkshop/DTOY/services/common"
	RegistryServiceClient "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/client"
	Chord "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/servant/dht"
	"github.com/TAULargeScaleWorkshop/DTOY/utils"
	"gopkg.in/yaml.v2"
)

var chord *Chord.Chord

func _GetRegistryAddresses() []string {
	var config ConfigBase.ConfigBase
	if len(os.Args) != 2 {
		utils.Logger.Printf("Expecting exactly one configuration file")
		os.Exit(1)
	}
	fileData, err := os.ReadFile(os.Args[1])
	if err != nil {
		utils.Logger.Fatalf("error reading file: %v", err)
	}

	err = yaml.Unmarshal(fileData, &config)
	if err != nil {
		utils.Logger.Fatalf("error unmarshaling data: %v", err)
	}

	if len(config.RegistryAddresses) == 0 {
		utils.Logger.Fatalf("No registry addresses found in configuration file")
	}
	return config.RegistryAddresses
}

func CreateChord(port int) {
	client := RegistryServiceClient.NewRegistryServiceClient(_GetRegistryAddresses())
	serviceNodes, err := client.Discover("CacheService")
	if err != nil {
		utils.Logger.Fatalf("Error discovering services: %v", err)
		return
	}

	for _, node := range serviceNodes {
		c := common.NewServiceClientBaseDirect(node)
		isRoot, err := c.IsRoot()
		if err != nil {
			utils.Logger.Printf("Error checking if node is root: %v", err)
			continue
		}
		if !isRoot {
			continue
		}

		utils.Logger.Printf("Joining chord %v at chord port: %v\n", "root_node", 1097)
		chord, err = Chord.JoinChord(fmt.Sprintf("CacheService%v", port), "root_node", 1097)
		if err != nil {
			utils.Logger.Fatalf("Error joining chord: %v", err)
		}
		return
	}

	utils.Logger.Printf("No root node found, creating new chord")
	utils.Logger.Printf("Root node name is: %v\n", "root_node")
	chord, err = Chord.NewChord("root_node", 1097)
	if err != nil {
		utils.Logger.Fatalf("Error creating new chord: %v", err)
	}
}

func GetPortFromNode(node string) string {
	// split by :
	s := strings.Split(node, ":")
	return s[len(s)-1]
}

func _CheckIfKeyInKeys(key string) (bool, error) {
	keys, err := chord.GetAllKeys()
	if err != nil {
		utils.Logger.Printf("Failed to get all keys: %v\n", err)
		return false, err
	}
	for _, k := range keys {
		if k == key {
			return true, nil
		}
	}
	return false, nil
}

func Get(key string) (string, error) {
	found, err := _CheckIfKeyInKeys(key)
	if err != nil || !found {
		return "", nil
	}
	res, err := chord.Get(key)
	if err != nil {
		return "", err
	}
	return res, nil
}

func Set(key string, value string) error {
	err := chord.Set(key, value)
	if err != nil {
		return err
	}
	return nil
}

func Delete(key string) error {
	found, err := _CheckIfKeyInKeys(key)
	if err != nil || !found {
		return nil
	}
	err = chord.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

func IsRoot() (bool, error) {
	val, err := chord.IsFirst()
	if err != nil {
		return false, err
	}
	return val, nil
}
