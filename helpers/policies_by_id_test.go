package helpers

import (
	"reflect"
	"testing"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"gopkg.in/mgo.v2/bson"
)

type DummyPolicyParams struct {
	PolicyID int
	Tags     []string
}

var (
	objectIds = map[int]bson.ObjectId{}
	once      = false
)

func GenerateDummyPolicy(param DummyPolicyParams) objects.Policy {
	// Generating random Object Ids
	if !once {
		for i := 1; i <= 5; i++ {
			objectIds[i] = bson.NewObjectId()
		}
		once = true
	}

	dummyPolicy := objects.Policy{}

	if param.PolicyID != 0 {
		dummyPolicy.MID = objectIds[param.PolicyID]
	}

	if len(param.Tags) > 0 {
		dummyPolicy.Tags = param.Tags
	}

	return dummyPolicy
}

func TestGetPoliciesByID(t *testing.T) {

	type args struct {
		totalPolicies      []objects.Policy
		wantedPoliciesByID []string
	}
	tests := []struct {
		name    string
		args    args
		want    []objects.Policy
		wantErr bool
	}{
		{
			name: "1 Policy with wanted ID",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				},
				wantedPoliciesByID: []string{objectIds[1].Hex()},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
			},
			wantErr: false,
		},
		{
			name: "1 Policy with wanted ID",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				},
				wantedPoliciesByID: []string{objectIds[1].Hex()},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
			},
			wantErr: false,
		},
		{
			name: "2 Policies with 2 wanted ID",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
				},
				wantedPoliciesByID: []string{objectIds[1].Hex(), objectIds[2].Hex()},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
			},
			wantErr: false,
		},
		{
			name: "Creating 2 policies, but only 1 is wanted",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
				},
				wantedPoliciesByID: []string{objectIds[2].Hex()},
			},
			want: []objects.Policy{
				GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
			},
			wantErr: false,
		},
		{
			name: "Creating 2 policies, but none is wanted",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 1}),
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
				},
				wantedPoliciesByID: []string{objectIds[3].Hex()},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Creating 1 policy, but 2 are wanted",
			args: args{
				totalPolicies: []objects.Policy{
					GenerateDummyPolicy(DummyPolicyParams{PolicyID: 2}),
				},
				wantedPoliciesByID: []string{objectIds[2].Hex(), objectIds[3].Hex()},
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPoliciesByID(tt.args.totalPolicies, tt.args.wantedPoliciesByID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPoliciesByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPoliciesByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
