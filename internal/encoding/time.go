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
// limitations under the License

package encoding

import (
	"time"

	"github.com/pkg/errors"
)

// EncodeTimeAscending encodes a time value, appends it to the supplied buffer,
// and returns the final buffer. The encoding is guaranteed to be ordered
// Such that if t1.Before(t2) then after EncodeTime(b1, t1), and
// EncodeTime(b2, t2), Compare(b1, b2) < 0. The time zone offset not
// included in the encoding.
func EncodeTimeAscending(b []byte, t time.Time) []byte {
	return encodeTime(b, t.Unix(), int64(t.Nanosecond()))
}

// EncodeTimeDescending is the descending version of EncodeTimeAscending.
func EncodeTimeDescending(b []byte, t time.Time) []byte {
	return encodeTime(b, ^t.Unix(), ^int64(t.Nanosecond()))
}

func encodeTime(b []byte, unix, nanos int64) []byte {
	// Read the unix absolute time. This is the absolute time and is
	// not time zone offset dependent.
	b = append(b, timeMarker)
	b = EncodeVarintAscending(b, unix)
	b = EncodeVarintAscending(b, nanos)
	return b
}

// DecodeTimeAscending decodes a time.Time value which was encoded using
// EncodeTime. The remainder of the input buffer and the decoded
// time.Time are returned.
func DecodeTimeAscending(b []byte) ([]byte, time.Time, error) {
	b, sec, nsec, err := decodeTime(b)
	if err != nil {
		return b, time.Time{}, err
	}
	return b, time.Unix(sec, nsec).UTC(), nil
}

// DecodeTimeDescending is the descending version of DecodeTimeAscending.
func DecodeTimeDescending(b []byte) ([]byte, time.Time, error) {
	b, sec, nsec, err := decodeTime(b)
	if err != nil {
		return b, time.Time{}, err
	}
	return b, time.Unix(^sec, ^nsec).UTC(), nil
}

func decodeTime(b []byte) (r []byte, sec int64, nsec int64, err error) {
	if PeekType(b) != Time {
		return nil, 0, 0, errors.Errorf("did not find marker")
	}
	b = b[1:]
	b, sec, err = DecodeVarintAscending(b)
	if err != nil {
		return b, 0, 0, err
	}
	b, nsec, err = DecodeVarintAscending(b)
	if err != nil {
		return b, 0, 0, err
	}
	return b, sec, nsec, nil
}
