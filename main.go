package main

import "nz.cloudwalker/wireguard-webadmin/repo/sqlite"

func main() {
	_, err := sqlite.NewSqliteRepository("file:test.db?cache=shared&mode=memory")
	if err != nil {
		panic(err)
	}
}
