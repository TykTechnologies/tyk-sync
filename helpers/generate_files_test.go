package helpers

import (
	"os"
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func TestGenerateApiFiles(t *testing.T) {

	// In this test, dir is not used since we don't want to create a directory with 777 permissions
	// So we're just storing the files in the "helpers" directory, and then, we delete them.

	type args struct {
		cleanApis     []objects.DBApiDefinition
		cleanPolicies []objects.Policy
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1 API",
			args: args{
				cleanApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
				},
				cleanPolicies: []objects.Policy{},
			},
			want:    []string{"api-1.json"},
			wantErr: false,
		},
		{
			name: "2 APIs",
			args: args{
				cleanApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
					GenerateDummyApi(DummyApiParams{ApiID: "2"}),
				},
				cleanPolicies: []objects.Policy{},
			},
			want:    []string{"api-1.json", "api-2.json"},
			wantErr: false,
		},
		{
			name: "Invalid character",
			args: args{
				cleanApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "/"}),
				},
				cleanPolicies: []objects.Policy{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "1 API with 1 non imported policy",
			args: args{
				cleanApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1", PolicyID: "1"}),
				},
				cleanPolicies: []objects.Policy{},
			},
			want:    []string{"api-1.json"},
			wantErr: false,
		},
		{
			name: "1 API with 1 imported policy",
			args: args{
				cleanApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1", PolicyID: objectIds[1].Hex()}),
				},
				cleanPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				},
			},
			want:    []string{"api-1.json"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Deleting files after test
			defer func() {
				for _, file := range tt.want {
					_ = os.RemoveAll(file)
				}
			}()

			got, err := GenerateApiFiles(tt.args.cleanApis, tt.args.cleanPolicies, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateApiFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateApiFiles() = %v, want %v", got, tt.want)
			}

			// Checking if files were created
			for _, file := range tt.want {
				_, err := os.Stat(file)
				if err != nil {
					t.Errorf("GenerateApiFiles() error = %v, wantErr %v", err, tt.wantErr)
				}

			}

		})
	}
}

func TestGeneratePolicyFiles(t *testing.T) {
	type args struct {
		cleanPolicies []objects.Policy
		cleanApis     []objects.DBApiDefinition
		dir           string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "1 policy",
			args: args{
				cleanPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				},
				cleanApis: []objects.DBApiDefinition{},
			},
			want:    []string{"policy-" + objectIds[1].Hex() + ".json"},
			wantErr: false,
		},
		{
			name: "2 policies",
			args: args{
				cleanPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
				},
				cleanApis: []objects.DBApiDefinition{},
			},
			want:    []string{"policy-" + objectIds[1].Hex() + ".json", "policy-" + objectIds[2].Hex() + ".json"},
			wantErr: false,
		}, {
			name: "1 policy with 1 non imported API",
			args: args{
				cleanPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1, ApiID: "1"}),
				},
				cleanApis: []objects.DBApiDefinition{},
			},
			want:    []string{"policy-" + objectIds[1].Hex() + ".json"},
			wantErr: false,
		},
		{
			name: "1 policy with 1 imported API",
			args: args{
				cleanPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1, ApiID: "1"}),
				},
				cleanApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "1"}),
				},
			},
			want:    []string{"policy-" + objectIds[1].Hex() + ".json"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Deleting files after test
			defer func() {
				for _, file := range tt.want {
					_ = os.RemoveAll(file)
				}
			}()

			got, err := GeneratePolicyFiles(tt.args.cleanPolicies, tt.args.cleanApis, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePolicyFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GeneratePolicyFiles() = %v, want %v", got, tt.want)
			}
			// Checking if files were created
			for _, file := range tt.want {
				_, err := os.Stat(file)
				if err != nil {
					t.Errorf("GenerateApiFiles() error = %v, wantErr %v", err, tt.wantErr)
				}

			}
		})
	}
}
