prepare:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/twitchtv/twirp/protoc-gen-twirp

api/rpc: prepare
	mkdir -p $@

gen: api/rpc api/twirp_rpc.proto
	env PATH=${PATH}:${GOPATH}/bin protoc --proto_path=api --twirp_out=$< --go_out=$< twirp_rpc.proto
