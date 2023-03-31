/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package consul

import (
	"reflect"
	"testing"
)

func TestSplitTags(t *testing.T) {
	type args struct {
		tags []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Regular tags",
			args: args{
				tags: []string{"k1:v1", "k2:v2"},
			},
			want: map[string]string{"k1": "v1", "k2": "v2"},
		},
		{
			name: "Some tags no values",
			args: args{
				tags: []string{"k1:", "k2:v2"},
			},
			want: map[string]string{"k1": "", "k2": "v2"},
		},
		{
			name: "Tags char splited,two elements be handled correctly",
			args: args{
				tags: []string{"k1:v1:vv1", "k2:v2"},
			},
			want: map[string]string{"k1": "v1:vv1", "k2": "v2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitTags(tt.args.tags); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
