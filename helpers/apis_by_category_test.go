package helpers

import (
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk/apidef"
)

type DummyApiParams struct {
	Category string
	ApiID    string
	Tags     []string
}

func GenerateDummyApi(params DummyApiParams) objects.DBApiDefinition {
	dummyApi := objects.DBApiDefinition{
		APIDefinition: &objects.APIDefinition{
			APIDefinition: apidef.APIDefinition{
				Name: "dummy-api",
			},
		},
	}

	if params.Category != "" {
		dummyApi.APIDefinition.Name += "#" + params.Category
	}

	if params.ApiID != "" {
		dummyApi.APIID = params.ApiID
	}

	if len(params.Tags) > 0 {
		dummyApi.APIDefinition.Tags = params.Tags
	}

	return dummyApi
}

func TestGetApisByCategory(t *testing.T) {
	type args struct {
		totalApis        []objects.DBApiDefinition
		wantedCategories []string
	}
	tests := []struct {
		name    string
		args    args
		want    []objects.DBApiDefinition
		wantErr bool
	}{
		{
			name: "1 API with wanted category",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
				},
				wantedCategories: []string{"wanted-category"},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
			},
			wantErr: false,
		},
		{
			name: "2 APIs with wanted category",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
					GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
				},
				wantedCategories: []string{"wanted-category"},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
				GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
			},
			wantErr: false,
		},
		{
			name: "Creating 2 APIS, but only one wanted",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
					GenerateDummyApi(DummyApiParams{Category: "non-wanted-category"}),
				},
				wantedCategories: []string{"wanted-category"},
			},
			want: []objects.DBApiDefinition{
				GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
			},
			wantErr: false,
		},
		{
			name: "Creating 1 API, but wanted 2",
			args: args{
				totalApis: []objects.DBApiDefinition{
					GenerateDummyApi(DummyApiParams{Category: "wanted-category"}),
				},
				wantedCategories: []string{"wanted-category", "wanted-category-2"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Creating 0 APIs",
			args: args{
				totalApis:        []objects.DBApiDefinition{},
				wantedCategories: []string{"wanted-category"},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetApisByCategory(tt.args.totalApis, tt.args.wantedCategories)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetApisByCategory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetApisByCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}
