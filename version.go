// Copyright (c) 2026-present, Nexthop Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
)

// buildInfo holds VCS metadata embedded by the Go toolchain at build time.
var buildInfo = readBuildInfo()

type versionInfo struct {
	Revision  string `json:"revision"`  // full git SHA1, or "unknown"
	BuildTime string `json:"buildTime"` // RFC3339 commit timestamp, or "unknown"
	Dirty     bool   `json:"dirty"`     // true if built from an unclean working tree
}

func readBuildInfo() versionInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return versionInfo{Revision: "unknown", BuildTime: "unknown"}
	}
	v := versionInfo{Revision: "unknown", BuildTime: "unknown"}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			v.Revision = s.Value
		case "vcs.time":
			v.BuildTime = s.Value
		case "vcs.modified":
			v.Dirty = s.Value == "true"
		}
	}
	return v
}

func handleVersion(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(buildInfo)
}

