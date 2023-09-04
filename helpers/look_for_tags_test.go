package helpers

import (
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func TestLookForTags(t *testing.T) {
	type args struct {
		totalPolicies []objects.Policy
		totalApis     []objects.DBApiDefinition
		wantedTags    []string
	}
	tests := []struct {
		name    string
		args    args
		want    []objects.Policy
		want1   []objects.DBApiDefinition
		wantErr bool
	}{
		{
			name: "1 API with wanted tag",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Tags: []string{"wanted-tag"}}),
				},
				wantedTags: []string{"wanted-tag"},
			},
			want:    nil,
			want1:   []objects.DBApiDefinition{GenerateDummyApi(DummyApiParams{Tags: []string{"wanted-tag"}})},
			wantErr: false,
		},
		{
			name: "1 Policy with wanted tag",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{Tags: []string{"wanted-tag"}}),
				},
				wantedTags: []string{"wanted-tag"},
			},
			want:    []objects.Policy{GenerateDummyPolicy(DummyPolicyParams{Tags: []string{"wanted-tag"}})},
			want1:   nil,
			wantErr: false,
		},
		{
			name: "1 Policy and 1 API with wanted tag",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{Tags: []string{"wanted-tag"}}),
				},
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Tags: []string{"wanted-tag"}}),
				},
				wantedTags: []string{"wanted-tag"},
			},
			want:    []objects.Policy{GenerateDummyPolicy(DummyPolicyParams{Tags: []string{"wanted-tag"}})},
			want1:   []objects.DBApiDefinition{GenerateDummyApi(DummyApiParams{Tags: []string{"wanted-tag"}})},
			wantErr: false,
		},
		{
			name: "1 Policy and 1 API with wanted tag, 1 API with unwanted tag",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{Tags: []string{"wanted-tag"}}),
				},
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Tags: []string{"wanted-tag"}}),
					GenerateDummyApi(DummyApiParams{Tags: []string{"unwanted-tag"}}),
				},
				wantedTags: []string{"wanted-tag"},
			},
			want:    []objects.Policy{GenerateDummyPolicy(DummyPolicyParams{Tags: []string{"wanted-tag"}})},
			want1:   []objects.DBApiDefinition{GenerateDummyApi(DummyApiParams{Tags: []string{"wanted-tag"}})},
			wantErr: false,
		},
		{
			name: "1 API with wanted tag, but 1 wanted tag is not present",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Tags: []string{"wanted-tag"}}),
				},
				wantedTags: []string{"wanted-tag", "wanted-tag-2"},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "1 Policy with wanted tag, but 1 wanted tag is not present",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{Tags: []string{"wanted-tag"}}),
				},
				wantedTags: []string{"wanted-tag", "wanted-tag-2"},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := LookForTags(tt.args.totalPolicies, tt.args.totalApis, tt.args.wantedTags)
			if (err != nil) != tt.wantErr {
				t.Errorf("LookForTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LookForTags() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("LookForTags() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
