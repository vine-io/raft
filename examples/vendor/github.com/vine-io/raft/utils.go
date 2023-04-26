// MIT License
//
// Copyright (c) 2023 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
