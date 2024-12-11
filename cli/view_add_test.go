// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileOrArgData(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		filePaths   []string
		expectedRes []string
		readFile    readFileFn
	}{
		{
			name:        "NoFile",
			args:        []string{"arg0", "arg1", "arg2"},
			filePaths:   []string{"", "", ""},
			expectedRes: []string{"arg0", "arg1", "arg2"},
		},
		{
			name:        "FileFirst",
			args:        []string{"arg1", "arg2"},
			filePaths:   []string{"file0", "", ""},
			expectedRes: []string{"file0", "arg1", "arg2"},
		},
		{
			name:        "FileMiddle",
			args:        []string{"arg0", "arg2"},
			filePaths:   []string{"", "file0", ""},
			expectedRes: []string{"arg0", "file0", "arg2"},
		},
		{
			name:        "FileLast",
			args:        []string{"arg0", "arg1"},
			filePaths:   []string{"", "", "file0"},
			expectedRes: []string{"arg0", "arg1", "file0"},
		},
		{
			name:        "FileFirstLast",
			args:        []string{"arg1"},
			filePaths:   []string{"file0", "", "file1"},
			expectedRes: []string{"file0", "arg1", "file1"},
		},
	}

	for _, test := range tests {
		fileReader := newMockReadFile()
		getData := newFileOrArgData(test.args, fileReader.Read)
		for i := range test.filePaths {
			res, err := getData.next(test.filePaths[i])
			require.NoError(t, err)
			require.Equal(t, test.expectedRes[i], res)
		}
	}
}

type mockReadFile struct {
	index int
}

func newMockReadFile() mockReadFile {
	return mockReadFile{}
}

func (f *mockReadFile) Read(string) ([]byte, error) {
	data := []byte(fmt.Sprintf("file%d", f.index))
	f.index += 1
	return data, nil
}
