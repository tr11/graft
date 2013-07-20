// Copyright 2013 Apcera Inc. All rights reserved.

package graft

import (
	"errors"
	"sync"
)

var mu sync.Mutex
var peers map[string]*Node

func init() {
	peers = make(map[string]*Node)
}

func mockPeerCount() int {
	mu.Lock()
	defer mu.Unlock()
	return len(peers)
}

func mockRegisterPeer(n *Node) {
	mu.Lock()
	defer mu.Unlock()
	peers[n.id] = n
}

func mockUnregisterPeer(id string) {
	mu.Lock()
	defer mu.Unlock()
	delete(peers, id)
}

type MockRpcDriver struct {
	mu   sync.Mutex
	node *Node

	// For testing
	shouldFailInit bool
	closeCalled    bool
	shouldFailComm bool
}

func NewMockRpc() *MockRpcDriver {
	return &MockRpcDriver{}
}

func (rpc *MockRpcDriver) Init(n *Node) error {
	if rpc.shouldFailInit {
		return errors.New("RPC Failed to Init")
	}
	mockRegisterPeer(n)
	rpc.node = n
	return nil
}

func (rpc *MockRpcDriver) Close() {
	rpc.closeCalled = true
	if rpc.node != nil {
		mockUnregisterPeer(rpc.node.id)
	}
}

func (rpc *MockRpcDriver) RequestVote(vr *VoteRequest) error {
	if rpc.isCommBlocked() {
		// Silent failure
		return nil
	}

	mu.Lock()
	defer mu.Unlock()
	for _, p := range peers {
		if p.id != rpc.node.id {
			p.VoteRequests <- vr
		}
	}
	return nil
}

func (rpc *MockRpcDriver) HeartBeat(hb *Heartbeat) error {
	if rpc.isCommBlocked() {
		// Silent failure
		return nil
	}

	mu.Lock()
	defer mu.Unlock()
	for _, p := range peers {
		if p.id != rpc.node.id {
			p.HeartBeats <- hb
		}
	}
	return nil
}

func (rpc *MockRpcDriver) SendVoteResponse(candidate string, vresp *VoteResponse) error {
	if rpc.isCommBlocked() {
		// Silent failure
		return nil
	}

	mu.Lock()
	defer mu.Unlock()
	if p := peers[candidate]; p != nil {
		p.VoteResponses <- vresp
	}
	return nil
}

func (rpc *MockRpcDriver) setCommBlocked(block bool) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.shouldFailComm = block
}

func (rpc *MockRpcDriver) isCommBlocked() bool {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	return rpc.shouldFailComm
}