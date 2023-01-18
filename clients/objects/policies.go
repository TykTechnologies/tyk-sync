package objects

import (
	"time"

	"github.com/TykTechnologies/graphql-go-tools/pkg/graphql"
	"gopkg.in/mgo.v2/bson"
)

type AccessSpec struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

// APILimit stores quota and rate limit on ACL level (per API)
type APILimit struct {
	Rate               float64 `json:"rate" bson:"rate"`
	Per                float64 `json:"per" bson:"per"`
	ThrottleInterval   float64 `json:"throttle_interval" bson:"throttle_interval"`
	ThrottleRetryLimit int     `json:"throttle_retry_limit" bson:"throttle_retry_limit"`
	MaxQueryDepth      int     `json:"max_query_depth" bson:"max_query_depth"`
	QuotaMax           int64   `json:"quota_max" bson:"quota_max"`
	QuotaRenews        int64   `json:"quota_renews" bson:"quota_renews"`
	QuotaRemaining     int64   `json:"quota_remaining" bson:"quota_remaining"`
	QuotaRenewalRate   int64   `json:"quota_renewal_rate" bson:"quota_renewal_rate"`
	SetByPolicy        bool    `json:"set_by_policy" bson:"set_by_policy"`
}

// AccessDefinition defines which versions of an API a key has access to
type AccessDefinition struct {
	APIName           string                  `json:"api_name" bson:"apiname"`
	APIID             string                  `json:"api_id" bson:"apiid"`
	Versions          []string                `json:"versions" bson:"versions"`
	AllowedURLs       []AccessSpec            `json:"allowed_urls" bson:"allowed_urls"` // mapped string MUST be a valid regex
	RestrictedTypes   []graphql.Type          `json:"restricted_types" bson:"restricted_types"`
	Limit             *APILimit               `json:"limit" bson:"limit"`
	FieldAccessRights []FieldAccessDefinition `json:"field_access_rights" bson:"field_access_rights"`

	AllowanceScope string `json:"allowance_scope" bson:"allowance_scope"`
}

type FieldAccessDefinition struct {
	TypeName  string      `json:"type_name" bson:"type_name"`
	FieldName string      `json:"field_name" bson:"field_name"`
	Limits    FieldLimits `json:"limits" bson:"limits"`
}

type FieldLimits struct {
	MaxQueryDepth int `json:"max_query_depth" bson:"max_query_depth"`
}

type Policy struct {
	MID                bson.ObjectId               `bson:"_id,omitempty" json:"_id"`
	ID                 string                      `bson:"id,omitempty" json:"id"`
	Name               string                      `bson:"name" json:"name"`
	OrgID              string                      `bson:"org_id" json:"org_id"`
	Rate               float64                     `bson:"rate" json:"rate"`
	Per                float64                     `bson:"per" json:"per"`
	QuotaMax           int64                       `bson:"quota_max" json:"quota_max"`
	QuotaRenewalRate   int64                       `bson:"quota_renewal_rate" json:"quota_renewal_rate"`
	ThrottleInterval   float64                     `bson:"throttle_interval" json:"throttle_interval"`
	ThrottleRetryLimit int                         `bson:"throttle_retry_limit" json:"throttle_retry_limit"`
	MaxQueryDepth      int                         `bson:"max_query_depth" json:"max_query_depth"`
	AccessRights       map[string]AccessDefinition `bson:"access_rights" json:"access_rights"`
	HMACEnabled        bool                        `bson:"hmac_enabled" json:"hmac_enabled"`
	Active             bool                        `bson:"active" json:"active"`
	IsInactive         bool                        `bson:"is_inactive" json:"is_inactive"`
	DateCreated        time.Time                   `bson:"date_created" json:"date_created"`
	Tags               []string                    `bson:"tags" json:"tags"`
	KeyExpiresIn       int64                       `bson:"key_expires_in" json:"key_expires_in"`
	Partitions         struct {
		Quota      bool `bson:"quota" json:"quota"`
		RateLimit  bool `bson:"rate_limit" json:"rate_limit"`
		Complexity bool `bson:"complexity" json:"complexity"`
		Acl        bool `bson:"acl" json:"acl"`
		PerAPI     bool `bson:"per_api" json:"per_api"`
	} `bson:"partitions" json:"partitions"`
	LastUpdated string                 `bson:"last_updated" json:"last_updated"`
	MetaData    map[string]interface{} `bson:"meta_data" json:"meta_data"`
}
