api/twirp:
	mkdir -p $@

gen: api/twirp api/twirp.proto
	env PATH=${PATH}:${GOPATH}/bin protoc --proto_path=api --twirp_out=$< --go_out=$< twirp.proto
