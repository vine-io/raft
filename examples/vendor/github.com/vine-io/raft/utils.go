package raft

import (
	"fmt"
	"hash/fnv"
	"strings"
)

func hash(text string) uint32 {
	h := fnv.New32()
	h.Write([]byte(text))
	return h.Sum32()
}

func peerUrl(peers []string, name string) (string, error) {
	for _, peer := range peers {
		if strings.HasPrefix(peer, name) {
			return strings.TrimPrefix(peer, name+"="), nil
		}
	}
	return "", fmt.Errorf("no matched")
}

func peerSplit(peer string) (string, string, error) {
	parts := strings.Split(peer, "=")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid peer")
	}
	return parts[0], parts[1], nil
}
