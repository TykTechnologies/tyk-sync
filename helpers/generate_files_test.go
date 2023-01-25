package helpers

import (
	"os"
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func TestGenerateApiFiles(t *testing.T) {

	err := os.MkdirAll("test", 0644)
	if err != nil {
		t.Error(err)
	}
	// defer os.RemoveAll(tempDir)

	type args struct {
		cleanApis     []objects.DBApiDefinition
		cleanPolicies []objects.Policy
		dir           string
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
				dir:           "test",
			},
			want:    []string{"test/1.json"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := GenerateApiFiles(tt.args.cleanApis, tt.args.cleanPolicies, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateApiFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateApiFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
