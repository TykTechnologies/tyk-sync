package helpers

import (
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func TestRemoveDuplicatesFromPolicies(t *testing.T) {
	type args struct {
		policies []objects.Policy
	}
	tests := []struct {
		name string
		args args
		want []objects.Policy
	}{
		{
			name: "1 policy",
			args: args{
				policies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
			},
		},
		{
			name: "2 policies with same ID",
			args: args{
				policies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
			},
		},
		{
			name: "2 policies with different ID",
			args: args{
				policies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
				},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
			},
		},
		{
			name: "3 policies and 2 with the same ID",
			args: args{
				policies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveDuplicatesFromPolicies(tt.args.policies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicatesFromPolicies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveDuplicatesFromApis(t *testing.T) {
	type args struct {
		apis []objects.DBApiDefinition
	}
	tests := []struct {
		name string
		args args
		want []objects.DBApiDefinition
	}{
		{
			name: "1 api",
			args: args{
				apis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
				},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{ApiID: "1"}),
			},
		},
		{
			name: "2 apis with same ID",
			args: args{
				apis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
				},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{ApiID: "1"}),
			},
		},
		{
			name: "2 apis with different ID",
			args: args{
				apis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
					GenerateDummyApi(DummyApiParams{ApiID: "2"}),
				},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{ApiID: "1"}),
				GenerateDummyApi(DummyApiParams{ApiID: "2"}),
			},
		},
		{
			name: "3 apis and 2 with the same ID",
			args: args{
				apis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
					GenerateDummyApi(DummyApiParams{ApiID: "2"}),
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
				},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{ApiID: "1"}),
				GenerateDummyApi(DummyApiParams{ApiID: "2"}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveDuplicatesFromApis(tt.args.apis); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveDuplicatesFromApis() = %v, want %v", got, tt.want)
			}
		})
	}
}
