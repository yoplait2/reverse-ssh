// reverseSSH - a lightweight ssh server with a reverse connection feature
// Copyright (C) 2021  Ferdinor <ferdinor@mailbox.org>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"github.com/gliderlabs/ssh"
)

// The following variables can be customised at compile time via -ldflags
// (see Makefile and build documentation). When built with `make`, RS_PASS
// defaults to a random hex string; all other variables fall back to the
// values shown here.
var (
	localPassword = "letmeinbrudipls" // password accepted for incoming SSH connections
	authorizedKey = ""                // authorized public key (empty = pubkey auth disabled)
	defaultShell  = "/bin/bash"       // shell spawned for interactive sessions
	version       = "1.3.0-dev"
	LUSER         = "reverse"  // username used when dialling home
	LHOST         = ""         // default target host (empty = bind/listen mode)
	LPORT         = "31337"    // listening port (bind) or target port (reverse)
	BPORT         = "8888"     // port bound on attacker side for reverse connections
	NOCLI         = ""         // non-empty value disables all CLI flag parsing
)

func main() {
	var (
		p              = setupParameters(NOCLI)
		forwardHandler = &ssh.ForwardedTCPHandler{}
		server         = ssh.Server{
			Handler:                       createSSHSessionHandler(p.shell),
			PasswordHandler:               createPasswordHandler(localPassword),
			PublicKeyHandler:              createPublicKeyHandler(authorizedKey),
			LocalPortForwardingCallback:   createLocalPortForwardingCallback(p.noShell),
			ReversePortForwardingCallback: createReversePortForwardingCallback(),
			SessionRequestCallback:        createSessionRequestCallback(p.noShell),
			ChannelHandlers: map[string]ssh.ChannelHandler{
				"direct-tcpip": ssh.DirectTCPIPHandler,
				"session":      ssh.DefaultSessionHandler,
				"rs-info":      createExtraInfoHandler(),
			},
			RequestHandlers: map[string]ssh.RequestHandler{
				"tcpip-forward":        forwardHandler.HandleSSHRequest,
				"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
			},
			SubsystemHandlers: map[string]ssh.SubsystemHandler{
				"sftp": createSFTPHandler(),
			},
		}
	)

	run(p, server)
}
