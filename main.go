package main

import (
	"log"
	"net"
	"nz.cloudwalker/wireguard-webadmin/wg"
)

//func syncRepositoryToWireguard(repository repo.Repository) error {
//	log.Println("Syncing repository to wireguard")
//	var result []repo.PeerInfo
//	var err error
//	if result, _, err = repository.ListPeers(0, 0); err != nil {
//		return err
//	}
//
//	log.Print("Got peers:", result)
//	return nil
//}

func main() {
	//repository, err := sqlite.NewSqliteRepository("file:test.db?cache=shared&mode=memory")
	//repository := repo.NewMemRepository()
	//
	//defer repository.Close()
	//
	//closed := make(chan interface{}, 1)
	//finished := make(chan interface{}, 1)
	//changes := make(chan interface{}, 1)
	//
	//repository.AddChangeNotification(changes)

	//go func() {
	//	defer func() {
	//		if e := recover(); e != nil {
	//			log.Println("Error syncing repository:", e)
	//		}
	//		finished <- nil
	//	}()
	//
	//	if err := syncRepositoryToWireguard(repository); err != nil {
	//		panic(err)
	//	}
	//
	//L:
	//	for {
	//		select {
	//		case _ = <-closed:
	//			{
	//				break L
	//			}
	//
	//		case _ = <-changes:
	//			{
	//				if err := syncRepositoryToWireguard(repository); err != nil {
	//					panic(err)
	//				}
	//			}
	//		}
	//	}
	//}()

	//go func() {
	//	key, _ := wgtypes.GeneratePrivateKey()
	//
	//	peers := []repo.PeerInfo{
	//		{
	//			PublicKey: key.PublicKey().String(),
	//			Name:      "First device",
	//		},
	//	}
	//
	//	if err := repository.UpdatePeers(peers); err != nil {
	//		panic(err)
	//	}
	//}()

	//httpApi, err := api.NewHttpApi(repository)
	//if err != nil {
	//	panic(err)
	//}
	//
	//if err := http.ListenAndServe("localhost:9090", cors.Default().Handler(httpApi)); err != nil {
	//	log.Printf("error serving http: %v", err)
	//}
	//
	//closed <- nil
	//<-finished

	client, err := wg.NewTunClient()
	if err != nil {
		panic(err)
	}

	defer client.Close()

	key, err := wg.NewRandom()
	if err != nil {
		panic(err)
	}

	_, address, err := net.ParseCIDR("192.168.30.1/24")
	if err != nil {
		panic(err)
	}

	device, err := client.Up("123", wg.DeviceConfig{
		Name:       "first device",
		PrivateKey: key,
		ListenPort: 12356,
		Address:    address,
	})

	if err != nil {
		panic(err)
	}

	log.Println("Got new device:", device)
}
