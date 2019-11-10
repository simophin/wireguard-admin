module nz.cloudwalker/wireguard-webadmin

go 1.13

require (
	github.com/jmoiron/sqlx v1.2.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/mattn/go-sqlite3 v1.11.0
	github.com/rs/cors v1.7.0
	github.com/vishvananda/netlink v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df // indirect
	golang.org/x/crypto v0.0.0-20191028145041-f83a4685e152
	golang.zx2c4.com/wireguard v0.0.20191012
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20191028205011-23406de29c08
)

replace golang.zx2c4.com/wireguard => ./wireguard-go
