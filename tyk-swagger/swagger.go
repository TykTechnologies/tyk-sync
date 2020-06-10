package tyk_swagger

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TykTechnologies/tyk/apidef"
	"github.com/ongoingio/urljoin"
	uuid "github.com/satori/go.uuid"
)

type DefinitionObjectFormatAST struct {
	Format string `json:"format"`
	Type   string `json:"type"`
}

type DefinitionObjectAST struct {
	Type       string                               `json:"type"`
	Required   []string                             `json:"required"`
	Properties map[string]DefinitionObjectFormatAST `json:"properties"`
}

type ResponseCodeObjectAST struct {
	Description string `json:"description"`
	Schema      struct {
		Items map[string]interface{} `json:"items"`
		Type  string                 `json:"type"`
	} `json:"schema"`
}

type PathMethodObject struct {
	Description string                           `json:"description"`
	OperationID string                           `json:"operationId"`
	Responses   map[string]ResponseCodeObjectAST `json:"responses"`
}

type PathItemObject struct {
	Get     PathMethodObject `json:"get"`
	Put     PathMethodObject `json:"put"`
	Post    PathMethodObject `json:"post"`
	Patch   PathMethodObject `json:"patch"`
	Options PathMethodObject `json:"options"`
	Delete  PathMethodObject `json:"delete"`
	Head    PathMethodObject `json:"head"`
}

type SwaggerAST struct {
	BasePath    string                         `json:"basePath"`
	Consumes    []string                       `json:"consumes"`
	Definitions map[string]DefinitionObjectAST `json:"definitions"`
	Host        string                         `json:"host"`
	Info        struct {
		Contact struct {
			Email string `json:"email"`
			Name  string `json:"name"`
			URL   string `json:"url"`
		} `json:"contact"`
		Description string `json:"description"`
		License     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"license"`
		TermsOfService string `json:"termsOfService"`
		Title          string `json:"title"`
		Version        string `json:"version"`
	} `json:"info"`
	Paths    map[string]PathItemObject `json:"paths"`
	Produces []string                  `json:"produces"`
	Schemes  []string                  `json:"schemes"`
	Swagger  string                    `json:"swagger"`
}

func (s *SwaggerAST) ReadString(asJson string) error {
	if err := json.Unmarshal([]byte(asJson), &s); err != nil {
		return err
	}
	return nil
}

func NewEmptyVersion() apidef.VersionInfo {
	// We need this because sometimes the schemas reject null values
	d := apidef.VersionInfo{}
	d.Paths.BlackList = make([]string, 0)
	d.Paths.WhiteList = make([]string, 0)
	d.Paths.Ignored = make([]string, 0)

	return d
}

func (s *SwaggerAST) ConvertIntoApiVersion(versionName string) (apidef.VersionInfo, error) {
	versionInfo := NewEmptyVersion()

	versionInfo.UseExtendedPaths = true
	vname := versionName
	if versionName == "" {
		vname = "Default"
	}

	versionInfo.Name = vname
	versionInfo.ExtendedPaths.TrackEndpoints = make([]apidef.TrackEndpointMeta, 0)

	if len(s.Paths) == 0 {
		return versionInfo, errors.New("no paths defined in swagger file")
	}

	for pathName, pathSpec := range s.Paths {
		newEndpointMeta := apidef.TrackEndpointMeta{}
		newEndpointMeta.Path = pathName

		// We just want the paths here, no mocks
		methods := map[string]PathMethodObject{
			"GET":     pathSpec.Get,
			"PUT":     pathSpec.Put,
			"POST":    pathSpec.Post,
			"HEAD":    pathSpec.Head,
			"PATCH":   pathSpec.Patch,
			"OPTIONS": pathSpec.Options,
			"DELETE":  pathSpec.Delete,
		}
		for methodName, m := range methods {
			// skip methods that are not defined
			if len(m.Responses) == 0 && m.Description == "" && m.OperationID == "" {
				continue
			}

			newEndpointMeta.Method = methodName
		}

		versionInfo.ExtendedPaths.TrackEndpoints = append(versionInfo.ExtendedPaths.TrackEndpoints, newEndpointMeta)
	}

	return versionInfo, nil
}

func newBlankDBDashDefinition() *apidef.APIDefinition {
	EmptyMW := apidef.MiddlewareSection{
		Pre:  make([]apidef.MiddlewareDefinition, 0),
		Post: make([]apidef.MiddlewareDefinition, 0),
	}
	return &apidef.APIDefinition{
		ConfigData:         map[string]interface{}{},
		ResponseProcessors: make([]apidef.ResponseProcessor, 0),
		AllowedIPs:         make([]string, 0),
		CustomMiddleware:   EmptyMW,
		Tags:               make([]string, 0),
	}
}

func CreateDefinitionFromSwagger(s *SwaggerAST, orgId string, versionName string) (*apidef.APIDefinition, error) {
	ad := newBlankDBDashDefinition()
	ad.Name = s.Info.Title
	ad.Active = true
	ad.UseKeylessAccess = true
	ad.APIID = uuid.NewV4().String()
	ad.OrgID = orgId

	ad.VersionDefinition.Key = "version"
	ad.VersionDefinition.Location = "header"
	ad.VersionData.Versions = make(map[string]apidef.VersionInfo)

	bp := s.BasePath
	if bp == "" {
		bp = fmt.Sprintf("/%v/", ad.APIID)
	}
	ad.Proxy.ListenPath = bp
	ad.Slug = bp

	h := s.Host
	if s.Host == "" {
		h = "unset.com"
	}

	trans := "http"
	if len(s.Schemes) > 0 {
		trans = s.Schemes[0]
	}

	ad.Proxy.StripListenPath = false
	host := fmt.Sprintf("%v://%v", trans, h)
	ad.Proxy.TargetURL = urljoin.Join(host, bp)

	versionData, err := s.ConvertIntoApiVersion(versionName)
	if err != nil {
		return nil, err
	}

	vname := versionName
	if vname == "" {
		vname = "Default"
		ad.VersionData.NotVersioned = true
	}

	ad.VersionData.Versions[vname] = versionData

	return ad, nil
}
