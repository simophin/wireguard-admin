all: gen_go

prepare:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/twitchtv/twirp/protoc-gen-twirp
	go get -u go.larrymyers.com/protoc-gen-twirp_typescript

GEN_FILES := api/rpc/twirp_rpc.twirp.go \
	api/rpc/twirp_rpc.pb.go \
	webapp/src/app/rpc/twirp.ts \
	webapp/src/app/rpc/twirp_rpc.ts

${GEN_FILES}: api/rpc webapp/src/app/rpc api/twirp_rpc.proto
	env PATH=${PATH}:${GOPATH}/bin protoc --proto_path=api --twirp_out=api/rpc \
		--go_out=api/rpc \
		--twirp_typescript_out=webapp/src/app/rpc \
		twirp_rpc.proto

gen_go: api/rpc/twirp_rpc.twirp.go api/rpc/twirp_rpc.pb.go

clean:
	rm -v ${GEN_FILES}