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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core/net"
	"github.com/sourcenetwork/defradb/http"
)

const commitHashMaxLength = 8

// Git info from build system. Public to be determined via the Makefile.
var (
	GoInfo        string
	GitRelease    string
	GitCommit     string
	GitCommitDate string
)

// defraVersion is the current version of DefraDB, its build information, and versions of components.
// It is serializable to JSON.
type defraVersion struct {
	Release    string `json:"release"`
	Commit     string `json:"commit"`
	CommitDate string `json:"commitdate"`
	GoInfo     string `json:"go"`

	VersionHTTPAPI string `json:"httpapi"`
	DocKeyVersions string `json:"dockeyversions"`
	NetProtocol    string `json:"netprotocol"`
}

// NewDefraVersion returns a defraVersion with normalized values.
func NewDefraVersion() (defraVersion, error) {
	dv := defraVersion{
		GoInfo:         strings.Replace(GoInfo, "go version go", "", 1),
		Release:        GitRelease,
		Commit:         GitCommit,
		CommitDate:     GitCommitDate,
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
	var commitHash string
	if len(dv.Commit) >= commitHashMaxLength {
		commitHash = dv.Commit[:commitHashMaxLength]
	}
	return fmt.Sprintf(
		`defradb %s (%s %s) built with Go %s`,
		dv.Release,
		commitHash,
		dv.CommitDate,
		dv.GoInfo,
	)
}

func (dv *defraVersion) StringFull() string {
	var commitHash string
	if len(dv.Commit) >= commitHashMaxLength {
		commitHash = dv.Commit[:commitHashMaxLength]
	}
	return fmt.Sprintf(
		`defradb %s (%s %s)
* HTTP API: %s
* P2P multicodec: %s
* DocKey versions: %s
* Go: %s`,
		dv.Release,
		commitHash,
		dv.CommitDate,
		dv.VersionHTTPAPI,
		dv.NetProtocol,
		dv.DocKeyVersions,
		dv.GoInfo,
	)
}
