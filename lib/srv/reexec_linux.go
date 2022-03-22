// Copyright 2021 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux
// +build linux

package srv

import (
	"os"
	"os/exec"
	"syscall"
)

// runningInQemuUser tries to determine if we're running inside of qemu-user;
// will err to the side of "not qemu" in case of internal errors, as that's not
// very common - and errors while we try to figure this out have a higher chance
// of happening on real systems anyway.
func runningInQemuUser() bool {
	// under qemu-user this is going to be a temporary file filled right before
	// returning from the syscall...
	statFile, err := os.Open("/proc/self/stat")
	if err != nil {
		return false
	}
	defer statFile.Close()

	// ...and regular files with content (/proc/self/stat is not empty) report a
	// nonzero size after fstat, whereas the actual procfs file reports zero
	statInfo, err := statFile.Stat()
	if err != nil {
		return false
	}

	return statInfo.Size() != 0
}

// procfsReexecOk indicates if it's safe to reexec by launching /proc/self/exe;
// this is true on regular Linux since at least kernel 2.2, but it's not true in
// qemu-user (6.2.0 and earlier, at time of writing), where we stick with
// launching os.Executable instead, and hope for the best.
//
// TODO(espadolini): if https://gitlab.com/qemu-project/qemu/-/issues/927 ends
// up being fixed in a way that lets us open and fexecve, do that instead
var procfsReexecOk = false

func init() {
	procfsReexecOk = !runningInQemuUser()
}

func reexecCommandOSTweaks(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = new(syscall.SysProcAttr)
	}
	// Linux only: when parent process (node) dies unexpectedly without
	// cleaning up child processes, send a signal for graceful shutdown
	// to children.
	cmd.SysProcAttr.Pdeathsig = syscall.SIGQUIT

	if procfsReexecOk {
		cmd.Path = "/proc/self/exe"
	}
}

func userCommandOSTweaks(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = new(syscall.SysProcAttr)
	}
	// Linux only: when parent process (this process) dies unexpectedly, kill
	// the child process instead of orphaning it.
	// SIGKILL because we don't control the child process and it could choose
	// to ignore other signals.
	cmd.SysProcAttr.Pdeathsig = syscall.SIGKILL
}
