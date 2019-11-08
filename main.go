package main

import (
	"github.com/rs/cors"
	"golang.zx2c4.com/wireguard/wgctrl"
	"log"
	"net/http"
	"nz.cloudwalker/wireguard-webadmin/api"
	"nz.cloudwalker/wireguard-webadmin/repo"
	"nz.cloudwalker/wireguard-webadmin/repo/wg"
)

func syncRepositoryToWireguard(repository repo.Repository) error {
	log.Println("Syncing repository to wireguard")
	var result []repo.PeerInfo
	var err error
	if result, _, err = repository.ListPeers(0, 0); err != nil {
		return err
	}

	log.Print("Got peers:", result)
	return nil
}

func main() {
	//repository, err := sqlite.NewSqliteRepository("file:test.db?cache=shared&mode=memory")
	client, err := wgctrl.New()
	if err != nil {
		panic(err)
	}
	repository, err := wg.NewWgCtlRepository(client)
	if err != nil {
		panic(err)
	}

	defer repository.Close()

	closed := make(chan interface{}, 1)
	finished := make(chan interface{}, 1)
	changes := make(chan interface{}, 1)

	repository.AddChangeNotification(changes)

	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Println("Error syncing repository:", e)
			}
			finished <- nil
		}()

		if err := syncRepositoryToWireguard(repository); err != nil {
			panic(err)
		}

	L:
		for {
			select {
			case _ = <-closed:
				{
					break L
				}

			case _ = <-changes:
				{
					if err := syncRepositoryToWireguard(repository); err != nil {
						panic(err)
					}
				}
			}
		}
	}()

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

	httpApi, err := api.NewHttpApi(repository)
	if err != nil {
		panic(err)
	}

	if err := http.ListenAndServe("localhost:9090", cors.Default().Handler(httpApi)); err != nil {
		log.Printf("error serving http: %v", err)
	}

	closed <- nil
	<-finished
}
