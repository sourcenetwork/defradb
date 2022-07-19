// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package version provides version information of DefraDB and components, and related facilities.
*/
package version

import (
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core/net"
)

const commitHashLength = 8

// Git info from build system.
var (
	GoInfo        string
	GitTag        string
	GitCommit     string
	GitCommitDate string
	GitBranch     string
)

// defraVersion is the current version of DefraDB, its build information, and versions of components.
// It is serializable to JSON.
type defraVersion struct {
	Release    string `json:"release"`
	Commit     string `json:"commit"`
	CommitDate string `json:"commitdate"`
	Branch     string `json:"branch"`
	GoInfo     string `json:"go"`

	VersionHTTPAPI string `json:"httpapi"`
	DocKeyVersions string `json:"dockeyversions"`
	NetProtocol    string `json:"netprotocol"`
}

// NewDefraVersion returns a DefraVersion with normalized values.
func NewDefraVersion() (defraVersion, error) {
	dv := defraVersion{
		GoInfo:         strings.Replace(GoInfo, "go version go", "", 1),
		Release:        GitTag,
		Commit:         GitCommit,
		CommitDate:     GitCommitDate,
		Branch:         GitBranch,
		VersionHTTPAPI: http.Version,
		NetProtocol:    string(net.Protocol),
	}
	var docKeyVersions []string
	for k, v := range client.ValidDocKeyVersions {
		if v {
			docKeyVersions = append(docKeyVersions, fmt.Sprintf("%x", k))
		}
	}
	dv.DocKeyVersions = strings.Join(docKeyVersions, ",")
	return dv, nil
}

func (dv *defraVersion) String() string {
	// short commit hash
	var commitHash strings.Builder
	for i, r := range dv.Commit {
		if i > commitHashLength {
			break
		}
		commitHash.WriteRune(r)
	}
	return fmt.Sprintf(
		`defradb %s (%s %s)
http api: %s
net protocol: %s
dockey versions: %s
go: %s`,
		dv.Release,
		commitHash.String(),
		dv.CommitDate,
		dv.VersionHTTPAPI,
		dv.NetProtocol,
		dv.DocKeyVersions,
		dv.GoInfo,
	)
}
