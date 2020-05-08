package objects

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type AccessSpec struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

type AccessDefinition struct {
	APIName     string       `bson:"apiname" json:"api_name"`
	APIID       string       `bson:"apiid" json:"api_id"`
	Versions    []string     `bson:"versions" json:"versions"`
	AllowedURLs []AccessSpec `bson:"allowed_urls"  json:"allowed_urls"` // mapped string MUST be a valid regex
}

type Policy struct {
	MID              bson.ObjectId               `bson:"_id,omitempty" json:"_id"`
	ID               string                      `bson:"id,omitempty" json:"id"`
	OrgID            string                      `bson:"org_id" json:"org_id"`
	Rate             float64                     `bson:"rate" json:"rate"`
	Per              float64                     `bson:"per" json:"per"`
	QuotaMax         int64                       `bson:"quota_max" json:"quota_max"`
	QuotaRenewalRate int64                       `bson:"quota_renewal_rate" json:"quota_renewal_rate"`
	AccessRights     map[string]AccessDefinition `bson:"access_rights" json:"access_rights"`
	HMACEnabled      bool                        `bson:"hmac_enabled" json:"hmac_enabled"`
	Active           bool                        `bson:"active" json:"active"`
	Name             string                      `bson:"name" json:"name"`
	IsInactive       bool                        `bson:"is_inactive" json:"is_inactive"`
	DateCreated      time.Time                   `bson:"date_created" json:"date_created"`
	Tags             []string                    `bson:"tags" json:"tags"`
	KeyExpiresIn     int64                       `bson:"key_expires_in" json:"key_expires_in"`
	Partitions       struct {
		Quota     bool `bson:"quota" json:"quota"`
		RateLimit bool `bson:"rate_limit" json:"rate_limit"`
		Acl       bool `bson:"acl" json:"acl"`
	} `bson:"partitions" json:"partitions"`
	LastUpdated string `bson:"last_updated" json:"last_updated"`
}

func (pol *Policy) FixPolicyAPIIDs(APIIDRelations map[string]string) {
	apiIDToRemove := []string{}
	for apiID, accessRights := range pol.AccessRights {
		newAPIID, found := APIIDRelations[apiID]
		if found {
			newAccessRights := accessRights
			newAccessRights.APIID = newAPIID
			pol.AccessRights[newAPIID] = newAccessRights
			apiIDToRemove = append(apiIDToRemove, apiID)
		}
	}

	for _, apiID := range apiIDToRemove {
		delete(pol.AccessRights, apiID)
	}
}
