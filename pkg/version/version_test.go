/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import (
	"testing"
)

func TestBinaryVersion(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			name: "happy case",
			arg:  "v0.0.15",
			want: "v0.0.15",
		},
		{
			name: "no ldflags",
			arg:  "v0.0.0-master+$Format:%H$",
			want: "v0.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BinaryVersion(tt.arg)
			if got != tt.want {
				t.Errorf("binaryVersion() got = %v, want %v", got, tt.want)
			}
		})
	}
}
