package tyk_vcs

type PublishAction string
type SpecType string

const (
	CREATE PublishAction = "create"
	UPDATE PublishAction = "update"
	ERROR  PublishAction = "error"

	TYPE_APIDEF SpecType = "apidef"
	TYPE_OAI    SpecType = "oas"
)

type APIInfo struct {
	File  string `json:"file"`
	APIID string `json:"api_id"`
	DBID  string `json:"db_id"`
	ORGID string `json:"org_id"`
	OAS   struct {
		OverrideTarget     string `json:"override_target"`
		OverrideListenPath string `json:"override_listen_path"`
		VersionName        string `json:"version_name"`
		StripListenPath    bool   `json:"strip_listen_path"`
	} `json:"oas"`
}

type PolicyInfo struct {
	File string `json:"file"`
	ID   string `json:"id"`
}

type TykSourceSpec struct {
	Type     SpecType     `json:"type"`
	Files    []APIInfo    `json:"files"`
	Policies []PolicyInfo `json:"policies"`
}
