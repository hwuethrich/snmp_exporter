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

//go:build unix

package scraper

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"
)

// SetTOSSocketOption returns a Control function that sets the IP_TOS socket option.
// The tos parameter should be a value between 0 and 255.
// A tos value of 0 means no TOS will be set.
func SetTOSSocketOption(tos int) func(network, address string, c syscall.RawConn) error {
	return func(network, address string, c syscall.RawConn) error {
		if tos == 0 {
			return nil
		}

		var opErr error
		err := c.Control(func(fd uintptr) {
			// Set IP_TOS for IPv4
			if network == "udp4" || network == "tcp4" {
				opErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_TOS, tos)
				if opErr != nil {
					opErr = fmt.Errorf("failed to set IP_TOS socket option: %w", opErr)
				}
			}
			// Set IPV6_TCLASS for IPv6
			if network == "udp6" || network == "tcp6" {
				opErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IPV6, unix.IPV6_TCLASS, tos)
				if opErr != nil {
					opErr = fmt.Errorf("failed to set IPV6_TCLASS socket option: %w", opErr)
				}
			}
		})
		if err != nil {
			return fmt.Errorf("failed to control raw connection: %w", err)
		}
		return opErr
	}
}
