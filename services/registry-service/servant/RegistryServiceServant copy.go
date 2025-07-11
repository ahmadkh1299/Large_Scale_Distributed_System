package RegistryServiceServant

// import (
// 	"fmt"
// 	"sync"
// 	"time"

// 	common "github.com/TAULargeScaleWorkshop/DTOY/services/common"
// 	dht "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/servant/dht"
// 	"github.com/TAULargeScaleWorkshop/DTOY/utils"
// )

// type ServiceNodesMap struct {
// 	lock     sync.Mutex
// 	cacheMap map[string][]string
// }

// var isAliveCheck map[string]int

// var serviceNodesMap *ServiceNodesMap

// var chord *dht.Chord

// func init() {
// 	serviceNodesMap = &ServiceNodesMap{cacheMap: make(map[string][]string)}
// 	isAliveCheck = make(map[string]int)
// }

// func CreateChord(baseListenPort int, listenPort int) {
// 	var err error
// 	if listenPort == baseListenPort {
// 		chord, err = dht.NewChord(fmt.Sprintf(":%v", listenPort), 1099)
// 		if err != nil {
// 			utils.Logger.Fatalf("Failed to create chord: %v\n", err)
// 		}
// 	} else {
// 		chord, err = dht.JoinChord(fmt.Sprintf(":%v", listenPort), fmt.Sprintf(":%v", baseListenPort), 1099)
// 		if err != nil {
// 			utils.Logger.Fatalf("Failed to join chord: %v\n", err)
// 		}
// 	}
// }

// func Discover(name string) []string {
// 	serviceNodesMap.lock.Lock()
// 	defer serviceNodesMap.lock.Unlock()
// 	if addresses, exists := serviceNodesMap.cacheMap[name]; exists {
// 		return addresses
// 	}
// 	return []string{}
// }

// func Register(name string, address string) {
// 	utils.Logger.Printf("Service %v registered on address: %v\n", name, address)
// 	serviceNodesMap.lock.Lock()
// 	if _, exists := serviceNodesMap.cacheMap[name]; exists {
// 		serviceNodesMap.cacheMap[name] = append(serviceNodesMap.cacheMap[name], address)
// 	} else {
// 		serviceNodesMap.cacheMap[name] = []string{address}
// 	}
// 	serviceNodesMap.lock.Unlock()
// }

// func Unregister(name string, address string) {
// 	serviceNodesMap.lock.Lock()
// 	if addresses, exists := serviceNodesMap.cacheMap[name]; exists {
// 		for i, addr := range addresses {
// 			if addr == address {
// 				serviceNodesMap.cacheMap[name] = append(addresses[:i], addresses[i+1:]...)
// 				break
// 			}
// 		}
// 		if len(serviceNodesMap.cacheMap[name]) == 0 {
// 			delete(serviceNodesMap.cacheMap, name)
// 		}
// 	}
// 	serviceNodesMap.lock.Unlock()
// }

// func CheckIsAliveEvery10Seconds() {
// 	var mut sync.Mutex
// 	done := true

// 	ticker := time.NewTicker(10 * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		// Wait for the ticker to fire
// 		<-ticker.C

// 		mut.Lock()
// 		if !done {
// 			mut.Unlock()
// 			continue
// 		}
// 		done = false
// 		mut.Unlock()
// 		go func() {
// 			utils.Logger.Printf("Started IsAlive check \n")
// 			CheckAllNodesStatus()
// 			mut.Lock()
// 			done = true
// 			mut.Unlock()
// 		}()
// 	}
// }

// func DeleteByValue(slice []string, x string) []string {
// 	if len(slice) == 0 {
// 		return slice
// 	}

// 	for i, v := range slice {
// 		if v == x {
// 			return append(slice[:i], slice[i+1:]...)
// 		}
// 	}
// 	return slice
// }

// func CheckAllNodesStatus() {
// 	var services []string
// 	serviceNodesMap.lock.Lock()

// 	// Iterate over the map and collect the keys
// 	for key := range serviceNodesMap.cacheMap {
// 		services = append(services, key)
// 	}

// 	serviceNodesMap.lock.Unlock()

// 	for i := range services {
// 		serviceNodesMap.lock.Lock()
// 		nodes, ok := serviceNodesMap.cacheMap[services[i]]
// 		serviceNodesMap.lock.Unlock()
// 		if !ok {
// 			continue
// 		}
// 		var nodesToDelete []string = make([]string, len(nodes))

// 		for j, nodeAddr := range nodes {
// 			client := common.NewServiceClientBaseDirect(nodeAddr)
// 			prevFailures, found := isAliveCheck[nodeAddr]
// 			if !found {
// 				isAliveCheck[nodeAddr] = 0
// 				prevFailures = 0
// 			}
// 			res, err := client.IsAlive(services[i])
// 			if err != nil || !res {
// 				utils.Logger.Printf("Node %v of service %v is not alive.\n", nodeAddr, services[i])
// 				prevFailures++
// 			} else {
// 				prevFailures = 0
// 			}

// 			isAliveCheck[nodeAddr] = prevFailures

// 			if prevFailures >= 2 {
// 				nodesToDelete[j] = nodeAddr
// 			}
// 		}

// 		for _, nodeAddr := range nodesToDelete {
// 			if nodeAddr == "" {
// 				continue
// 			}
// 			serviceNodesMap.lock.Lock()
// 			serviceNodesMap.cacheMap[services[i]] = DeleteByValue(serviceNodesMap.cacheMap[services[i]], nodeAddr)
// 			serviceNodesMap.lock.Unlock()
// 		}

// 		serviceNodesMap.lock.Lock()
// 		if len(serviceNodesMap.cacheMap[services[i]]) == 0 {
// 			delete(serviceNodesMap.cacheMap, services[i])
// 		}
// 		utils.Logger.Printf("Service type: %v has the following services %v\n", services[i], serviceNodesMap.cacheMap[services[i]])
// 		serviceNodesMap.lock.Unlock()
// 	}
// }
