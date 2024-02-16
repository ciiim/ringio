#!/bin/bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./ringio/fspb/nodeservice.proto
protoc --go_out=. --go_opt=paths=source_relative ./ringio/fspb/chunkinfo.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./ringio/fspb/chunkservice.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./ringio/fspb/treefsservice.proto
protoc --go_out=. --go_opt=paths=source_relative ./ringio/fspb/error.proto