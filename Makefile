all: gen

${GOPATH}/bin/protoc-gen-go:
	go get -u github.com/golang/protobuf/protoc-gen-go

${GOPATH}/bin/protoc-gen-twirp:
	go get -u github.com/twitchtv/twirp/protoc-gen-twirp

${GOPATH}/protoc-gen-twirp_typescript:
	go get -u go.larrymyers.com/protoc-gen-twirp_typescript

.PHONY prepare: protoc ${GOPATH}/bin/protoc-gen-go ${GOPATH}/bin/protoc-gen-twirp ${GOPATH}/protoc-gen-twirp_typescript

PROTO_FILES := api/twirp_rpc.proto
GO_OUT_DIR := api/rpc
TS_OUT_DIR := webapp/src/app/rpc
GEN_FILES := ${GO_OUT_DIR}/twirp_rpc.twirp.go \
	${GO_OUT_DIR}/twirp_rpc.pb.go \
	${TS_OUT_DIR}/twirp.ts \
	${TS_OUT_DIR}/nz.cloudwalker.ts

${GO_OUT_DIR} ${TS_OUT_DIR}:
	mkdir -p $@

${GEN_FILES}: ${PROTO_FILES} ${GO_OUT_DIR} ${TS_OUT_DIR}
	env PATH=${PATH}:${GOPATH}/bin protoc --proto_path=$(<D) \
		--twirp_out=${GO_OUT_DIR} \
		--go_out=${GO_OUT_DIR} \
		--twirp_typescript_out=${TS_OUT_DIR} \
		$<

gen: ${GEN_FILES}

clean:
	rm -v ${GEN_FILES}