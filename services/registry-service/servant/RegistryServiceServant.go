package RegistryServiceServant

import (
	"fmt"
	"strings"
	"sync"
	"time"

	common "github.com/TAULargeScaleWorkshop/DTOY/services/common"
	dht "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/servant/dht"
	"github.com/TAULargeScaleWorkshop/DTOY/utils"
)

type ServiceNodesMap struct {
	lock     sync.Mutex
	cacheMap map[string][]string
}

var isAliveCheck map[string]int

var serviceNodesMap *ServiceNodesMap

var chord *dht.Chord

func init() {
	serviceNodesMap = &ServiceNodesMap{cacheMap: make(map[string][]string)}
	isAliveCheck = make(map[string]int)
}

func CreateChord(baseListenPort int, listenPort int) {
	var err error
	if listenPort == baseListenPort {
		chord, err = dht.NewChord(fmt.Sprintf(":%v", listenPort), 1098)
		if err != nil {
			utils.Logger.Fatalf("Failed to create chord: %v\n", err)
		}
	} else {
		chord, err = dht.JoinChord(fmt.Sprintf(":%v", listenPort), fmt.Sprintf(":%v", baseListenPort), 1098)
		if err != nil {
			utils.Logger.Fatalf("Failed to join chord: %v\n", err)
		}
	}
}

func EncodeStringArray(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return strings.Join(arr, ";")
}

func DecodeStringArray(str string) []string {
	if str == "" {
		return nil
	}
	return strings.Split(str, ";")
}

func CheckIfKeyInKeysAndSet(key string) {
	keys, err := chord.GetAllKeys()
	if err != nil {
		utils.Logger.Printf("Failed to get all keys: %v\n", err)
		return
	}
	for _, k := range keys {
		if k == key {
			return
		}
	}
	chord.Set(key, "")
}

func CheckIfKeyInKeysNoSet(key string) (bool, error) {
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

func Discover(name string) ([]string, error) {
	found, err := CheckIfKeyInKeysNoSet(name)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	res, err := chord.Get(name)
	if err != nil {
		return nil, err
	}
	return DecodeStringArray(res), nil
}

func Register(name string, address string) error {
	utils.Logger.Printf("Service %v registered on address: %v\n", name, address)
	CheckIfKeyInKeysAndSet(name)
	res, err := chord.Get(name)
	if err != nil {
		return err
	}
	addresses := DecodeStringArray(res)
	if addresses == nil {
		addresses = []string{}
	}
	addresses = append(addresses, address)
	err = chord.Set(name, EncodeStringArray(addresses))
	if err != nil {
		utils.Logger.Printf("Failed to register service %v on address %v\n", name, address)
		return err
	}
	return nil
}

func Unregister(name string, address string) error {
	res, err := chord.Get(name)
	if err != nil {
		return err
	}
	addresses := DecodeStringArray(res)
	if addresses == nil {
		return nil
	}
	for i, addr := range addresses {
		if addr == address {
			addresses = append(addresses[:i], addresses[i+1:]...)
			break
		}
	}
	if len(addresses) == 0 {
		chord.Delete(name)
	} else {
		chord.Set(name, EncodeStringArray(addresses))
	}
	return nil
}

func CheckIsAliveEvery10Seconds() {
	var mut sync.Mutex
	done := true

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		// Wait for the ticker to fire
		<-ticker.C

		mut.Lock()
		if !done {
			mut.Unlock()
			continue
		}
		done = false
		mut.Unlock()
		go func() {
			utils.Logger.Printf("Started IsAlive check \n")
			CheckAllNodesStatus()
			mut.Lock()
			done = true
			mut.Unlock()
		}()
	}
}

func DeleteByValue(slice []string, x string) []string {
	if len(slice) == 0 {
		return slice
	}

	for i, v := range slice {
		if v == x {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func CheckAllNodesStatus() {

	services, err := chord.GetAllKeys()
	if err != nil {
		utils.Logger.Printf("Failed to get all keys: %v\n", err)
		return
	}

	for i := range services {
		nodes, err := chord.Get(services[i])
		if err != nil {
			utils.Logger.Printf("Failed to get nodes for service %v: %v\n", services[i], err)
			continue
		}
		nodesArr := DecodeStringArray(nodes)
		if nodesArr == nil {
			continue
		}

		var nodesToDelete []string = make([]string, len(nodesArr))

		for j, nodeAddr := range nodesArr {
			client := common.NewServiceClientBaseDirect(nodeAddr)
			prevFailures, found := isAliveCheck[nodeAddr]
			if !found {
				isAliveCheck[nodeAddr] = 0
				prevFailures = 0
			}
			res, err := client.IsAlive(services[i])
			if err != nil || !res {
				utils.Logger.Printf("Node %v of service %v is not alive.\n", nodeAddr, services[i])
				prevFailures++
			} else {
				prevFailures = 0
			}

			isAliveCheck[nodeAddr] = prevFailures

			if prevFailures >= 2 {
				nodesToDelete[j] = nodeAddr
			}
		}

		for _, nodeAddr := range nodesToDelete {
			if nodeAddr == "" {
				continue
			}
			n, err := chord.Get(services[i])
			if err != nil {
				utils.Logger.Printf("Failed to get nodes for service %v: %v\n", services[i], err)
				continue
			}
			if n == "" {
				continue
			}

			nodesArr = DeleteByValue(DecodeStringArray(n), nodeAddr)
			if len(nodesArr) == 0 {
				chord.Delete(services[i])
			} else {
				err = chord.Set(services[i], EncodeStringArray(nodesArr))
				if err != nil {
					utils.Logger.Printf("Failed to delete node %v from service %v: %v\n", nodeAddr, services[i], err)
				}
			}
		}
		utils.Logger.Printf("Service type: %v has the following services %v\n", services[i], nodesArr)
	}
}
