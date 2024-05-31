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
	File  string `json:"file,omitempty"`
	APIID string `json:"api_id,omitempty"`
	DBID  string `json:"db_id,omitempty"`
	ORGID string `json:"org_id,omitempty"`
	OAS   struct {
		OverrideTarget     string `json:"override_target,omitempty"`
		OverrideListenPath string `json:"override_listen_path,omitempty"`
		VersionName        string `json:"version_name,omitempty"`
		StripListenPath    bool   `json:"strip_listen_path,omitempty"`
	} `json:"oas,omitempty"`
}

type PolicyInfo struct {
	File string `json:"file,omitempty"`
	ID   string `json:"id,omitempty"`
}

type AssetsInfo struct {
	File string `json:"file,omitempty"`
	ID   string `json:"id,omitempty"`
}

type TykSourceSpec struct {
	Type     SpecType     `json:"type,omitempty"`
	Files    []APIInfo    `json:"files,omitempty"`
	Policies []PolicyInfo `json:"policies,omitempty"`
	Assets   []AssetsInfo `json:"assets,omitempty"`
}
