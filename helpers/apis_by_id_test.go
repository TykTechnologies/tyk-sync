package helpers

import (
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func TestGetApisByID(t *testing.T) {
	type args struct {
		totalApis      []objects.DBApiDefinition
		wantedApisByID []string
	}
	tests := []struct {
		name    string
		args    args
		want    []objects.DBApiDefinition
		wantErr bool
	}{
		{
			name: "1 API with wanted ID",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
				},
				wantedApisByID: []string{"wanted-api-id"},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
			},
			wantErr: false,
		},
		{
			name: "2 API with 2 wanted ID",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id-2"}),
				},
				wantedApisByID: []string{"wanted-api-id", "wanted-api-id-2"},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
				GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id-2"}),
			},
			wantErr: false,
		},
		{
			name: "Creating 2 APIs, only 1 is wanted",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id-2"}),
				},
				wantedApisByID: []string{"wanted-api-id"},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
			},
			wantErr: false,
		},
		{
			name: "Creating 2 APIs, but wanting 1 that doesn't exist",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id-2"}),
				},
				wantedApisByID: []string{"wanted-api-id-3"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Creating 2 APIs, wanting 1 that doesn't exist and 1 that does",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id"}),
					GenerateDummyApi(DummyApiParams{ApiID: "wanted-api-id-2"}),
				},
				wantedApisByID: []string{"wanted-api-id-2", "wanted-api-id-3"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetApisByID(tt.args.totalApis, tt.args.wantedApisByID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetApisByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetApisByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
