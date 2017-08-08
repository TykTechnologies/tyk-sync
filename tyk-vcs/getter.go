package tyk_vcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TykTechnologies/tyk-git/tyk-swagger"
	"github.com/TykTechnologies/tyk/apidef"
	"gopkg.in/src-d/go-billy.v3"
	"gopkg.in/src-d/go-billy.v3/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"io/ioutil"
)

type GitGetter struct {
	repo      string
	branch    string
	key       []byte
	publisher Publisher

	r  *git.Repository
	fs billy.Filesystem
}

func NewGGetter(repo, branch string, key []byte, pub Publisher) (*GitGetter, error) {
	gh := &GitGetter{
		repo:      repo,
		branch:    branch,
		key:       key,
		fs:        memfs.New(),
		publisher: pub,
	}

	return gh, nil
}

func (gg *GitGetter) FetchRepo() error {
	// TODO: Auth support
	r, err := git.Clone(memory.NewStorage(), gg.fs, &git.CloneOptions{
		URL:           gg.repo,
		ReferenceName: plumbing.ReferenceName(gg.branch),
		SingleBranch:  true,
	})

	if err != nil {
		return err
	}

	gg.r = r

	return nil
}

func (gg *GitGetter) FetchTykSpec() (*TykSourceSpec, error) {
	if gg.r == nil {
		return nil, errors.New("No repository in memory, fetch repo first")
	}

	specFile, err := gg.fs.Open(".tyk.json")
	if err != nil {
		return nil, err
	}

	rawSpec, err := ioutil.ReadAll(specFile)
	if err != nil {
		return nil, err
	}

	ts := TykSourceSpec{}
	err = json.Unmarshal(rawSpec, &ts)
	if err != nil {
		return nil, err
	}

	return &ts, nil
}

func (gg *GitGetter) FetchAPIDef(spec *TykSourceSpec) ([]apidef.APIDefinition, error) {
	switch spec.Type {
	case TYPE_APIDEF:
		return gg.fetchAPIDefinitionsDirect(spec)
	case TYPE_OAI:
		return gg.fetchAPIDefinitionsFromOAI(spec)
	default:
		return nil, fmt.Errorf("Type must be '%v or '%v'", TYPE_APIDEF, TYPE_OAI)
	}

	return nil, nil
}

func (gg *GitGetter) fetchAPIDefinitionsDirect(spec *TykSourceSpec) ([]apidef.APIDefinition, error) {
	if gg.r == nil {
		return nil, errors.New("No repository in memory, fetch repo first")
	}

	defNames := spec.Files
	if len(spec.Files) == 0 {
		defNames = append(defNames, "api_definition.json")
	}

	defs := make([]apidef.APIDefinition, len(defNames))
	for i, defName := range defNames {
		defFile, err := gg.fs.Open(defName)
		if err != nil {
			return nil, err
		}

		rawDef, err := ioutil.ReadAll(defFile)
		if err != nil {
			return nil, err
		}

		ad := apidef.APIDefinition{}
		err = json.Unmarshal(rawDef, &ad)
		if err != nil {
			return nil, err
		}

		if spec.Meta.APIID != "" {
			ad.APIID = spec.Meta.APIID
		}

		if spec.Meta.ORGID != "" {
			ad.OrgID = spec.Meta.ORGID
		}

		defs[i] = ad
	}

	return defs, nil
}

func (gg *GitGetter) fetchAPIDefinitionsFromOAI(spec *TykSourceSpec) ([]apidef.APIDefinition, error) {

	if gg.r == nil {
		return nil, errors.New("No repository in memory, fetch repo first")
	}

	oaiNames := spec.Files
	if len(spec.Files) == 0 {
		oaiNames = append(oaiNames, "swagger.json")
	}

	defs := make([]apidef.APIDefinition, len(oaiNames))

	for i, oaiName := range(oaiNames) {
		oaiFile, err := gg.fs.Open(oaiName)
		if err != nil {
			return nil, err
		}

		rawData, err := ioutil.ReadAll(oaiFile)
		if err != nil {
			return nil, err
		}

		oai := tyk_swagger.SwaggerAST{}
		err = json.Unmarshal(rawData, &oai)
		if err != nil {
			return nil, err
		}

		ad, err := tyk_swagger.CreateDefinitionFromSwagger(&oai, spec.Meta.ORGID, spec.Meta.OAS.VersionName)
		if err != nil {
			return nil, err
		}

		if spec.Meta.APIID != "" {
			ad.APIID = spec.Meta.APIID
		}

		if spec.Meta.OAS.OverrideListenPath != "" {
			ad.Proxy.ListenPath = spec.Meta.OAS.OverrideListenPath
		}

		if spec.Meta.OAS.OverrideTarget != "" {
			ad.Proxy.TargetURL = spec.Meta.OAS.OverrideTarget
		}

		if spec.Meta.OAS.StripListenPath {
			ad.Proxy.StripListenPath = true
		}

		defs[i] = *ad
	}

	return defs, nil
}

func (gg *GitGetter) Create(apiDef *apidef.APIDefinition) (string, error) {
	return gg.publisher.Create(apiDef)
}

func (gg *GitGetter) Update(id string, apiDef *apidef.APIDefinition) error {
	return gg.publisher.Update(id, apiDef)
}
