#!/bin/bash
protoc --go_out=. --go_opt=paths=source_relative ./internal/fs/fspb/fileinfo.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./internal/fs/fspb/peer_grpc.proto