// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

type ByteSize uint64

const (
	B   ByteSize = 1
	KiB          = B << 10
	MiB          = KiB << 10
	GiB          = MiB << 10
	TiB          = GiB << 10
	PiB          = TiB << 10
)

// UnmarshalText calls Set on ByteSize with the given text
func (bs *ByteSize) UnmarshalText(text []byte) error {
	return bs.Set(string(text))
}

// String returns the string formatted output of ByteSize
func (bs *ByteSize) String() string {
	const unit = 1024
	bsInt := int64(*bs)
	if bsInt < unit {
		return fmt.Sprintf("%d", bsInt)
	}
	div, exp := int64(unit), 0
	for n := bsInt / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%d%ciB", bsInt/div, "KMGTP"[exp])
}

// Type returns the type as a string.
func (bs *ByteSize) Type() string {
	return "ByteSize"
}

// Set parses a string into ByteSize
func (bs *ByteSize) Set(s string) error {
	digitString := ""
	unit := ""
	for _, char := range s {
		if unicode.IsDigit(char) {
			digitString += string(char)
		} else {
			unit += string(char)
		}
	}
	digits, err := strconv.Atoi(digitString)
	if err != nil {
		return NewErrUnableToParseByteSize(err)
	}

	switch strings.ToUpper(strings.Trim(unit, " ")) {
	case "B":
		*bs = ByteSize(digits) * B
	case "KB", "KIB":
		*bs = ByteSize(digits) * KiB
	case "MB", "MIB":
		*bs = ByteSize(digits) * MiB
	case "GB", "GIB":
		*bs = ByteSize(digits) * GiB
	case "TB", "TIB":
		*bs = ByteSize(digits) * TiB
	case "PB", "PIB":
		*bs = ByteSize(digits) * PiB
	default:
		*bs = ByteSize(digits)
	}

	return nil
}

// expandHomeDir expands paths if they were passed in as `~` rather than `${HOME}`
// converts `~/.defradb/certs/server.crt` to `/home/username/.defradb/certs/server.crt`.
func expandHomeDir(path *string) error {
	if *path == "~" {
		return ErrPathCannotBeHomeDir
	} else if strings.HasPrefix(*path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return NewErrUnableToExpandHomeDir(err)
		}

		// Use strings.HasPrefix so we don't match paths like "/x/~/x/"
		*path = filepath.Join(homeDir, (*path)[2:])
	}

	return nil
}

func isLowercaseAlpha(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}

func parseKV(kv string) ([]string, error) {
	parsedKV := strings.Split(kv, "=")
	if len(parsedKV) != 2 {
		return nil, NewErrNotProvidedAsKV(kv)
	}
	if parsedKV[0] == "" || parsedKV[1] == "" {
		return nil, NewErrNotProvidedAsKV(kv)
	}
	return parsedKV, nil
}
