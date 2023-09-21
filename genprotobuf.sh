#!/bin/bash
protoc --go_out=. --go_opt=paths=source_relative ./internal/dfs/fspb/fileinfo.proto
protoc --go_out=. --go_opt=paths=source_relative ./internal/dfs/fspb/error.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./internal/dfs/fspb/peer_grpc.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./internal/dfs/fspb/file_grpc.proto