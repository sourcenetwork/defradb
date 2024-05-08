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
	"bytes"
	"testing"
)

func TestEncodeDecodeNull(t *testing.T) {
	const hello = "hello"

	buf := EncodeNullAscending([]byte(hello))
	expected := []byte(hello + "\x00")
	if !bytes.Equal(expected, buf) {
		t.Fatalf("expected %q, but found %q", expected, buf)
	}

	if remaining, isNull := DecodeIfNull([]byte(hello)); isNull {
		t.Fatalf("expected isNull=false, but found isNull=%v", isNull)
	} else if hello != string(remaining) {
		t.Fatalf("expected %q, but found %q", hello, remaining)
	}

	if remaining, isNull := DecodeIfNull([]byte("\x00" + hello)); !isNull {
		t.Fatalf("expected isNull=true, but found isNull=%v", isNull)
	} else if hello != string(remaining) {
		t.Fatalf("expected %q, but found %q", hello, remaining)
	}
}
