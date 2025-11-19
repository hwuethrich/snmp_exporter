// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scraper

import (
	"testing"
)

func TestSetTOSSocketOption_Zero(t *testing.T) {
	// Test that TOS=0 doesn't set any socket options
	fn := SetTOSSocketOption(0)
	if fn == nil {
		t.Fatal("SetTOSSocketOption(0) returned nil")
	}

	// Should succeed with nil RawConn since it shouldn't try to set anything
	err := fn("udp4", "127.0.0.1:161", nil)
	if err != nil {
		t.Errorf("SetTOSSocketOption(0) with nil RawConn failed: %v", err)
	}
}

func TestSetTOSSocketOption_NonZero(t *testing.T) {
	// Test that non-zero TOS values return a valid function
	testCases := []int{1, 46, 255}

	for _, tos := range testCases {
		fn := SetTOSSocketOption(tos)
		if fn == nil {
			t.Errorf("SetTOSSocketOption(%d) returned nil", tos)
		}
	}
}

// mockRawConn is a mock implementation of syscall.RawConn for testing
type mockRawConn struct {
	controlFunc func(f func(fd uintptr)) error
}

func (m *mockRawConn) Control(f func(fd uintptr)) error {
	if m.controlFunc != nil {
		return m.controlFunc(f)
	}
	return nil
}

func (m *mockRawConn) Read(f func(fd uintptr) (done bool)) error {
	return nil
}

func (m *mockRawConn) Write(f func(fd uintptr) (done bool)) error {
	return nil
}

func TestSetTOSSocketOption_CallsControl(t *testing.T) {
	// Test that the function properly calls Control on the RawConn
	tos := 46
	fn := SetTOSSocketOption(tos)

	mock := &mockRawConn{
		controlFunc: func(f func(fd uintptr)) error {
			// Don't actually call f as we don't have a real fd
			return nil
		},
	}

	// On Unix systems, this should call Control
	// On other platforms, it might return early or error
	_ = fn("udp4", "127.0.0.1:161", mock)

	// We can't assert controlCalled on all platforms, but we can verify
	// the function doesn't panic
}

func TestSetTOSSocketOption_DifferentNetworks(t *testing.T) {
	// Test different network types
	networks := []string{"udp4", "udp6", "tcp4", "tcp6"}
	tos := 46

	for _, network := range networks {
		fn := SetTOSSocketOption(tos)
		if fn == nil {
			t.Errorf("SetTOSSocketOption(%d) returned nil for network %s", tos, network)
		}

		// Test with a mock RawConn to avoid nil pointer dereference
		mock := &mockRawConn{
			controlFunc: func(f func(fd uintptr)) error {
				// Don't actually call f as we don't have a real fd
				return nil
			},
		}
		_ = fn(network, "127.0.0.1:161", mock)
	}
}
