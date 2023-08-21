#!/bin/bash
protoc --go_out=. --go_opt=paths=source_relative fileinfo.proto
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative peer_grpc.proto