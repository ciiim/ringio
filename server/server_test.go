package server

import (
	"testing"
)

func TestServer_BeginStoreFile(t *testing.T) {
	server := NewServer("test", "test", "127.0.0.1", "9632")
	server.BeginStoreFile("test", "test", "test", "test", 1)
}
