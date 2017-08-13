package tyk_vcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TykTechnologies/tyk-git/clients/objects"
	"github.com/TykTechnologies/tyk-git/tyk-swagger"
	"github.com/TykTechnologies/tyk/apidef"
	"gopkg.in/mgo.v2/bson"
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
		fmt.Println(".tyk.json")
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
	defs := make([]apidef.APIDefinition, len(defNames))
	for i, defInfo := range defNames {
		defFile, err := gg.fs.Open(defInfo.File)
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

		if defInfo.APIID != "" {
			ad.APIID = defInfo.APIID
		}

		if defInfo.DBID != "" {
			ad.Id = bson.ObjectIdHex(defInfo.DBID)
		}

		if defInfo.ORGID != "" {
			ad.OrgID = defInfo.ORGID
		}

		defs[i] = ad
	}

	fmt.Printf("Fetched %v definitions\n", len(defs))

	return defs, nil
}

func (gg *GitGetter) fetchAPIDefinitionsFromOAI(spec *TykSourceSpec) ([]apidef.APIDefinition, error) {

	if gg.r == nil {
		return nil, errors.New("No repository in memory, fetch repo first")
	}

	oaiNames := spec.Files
	defs := make([]apidef.APIDefinition, len(oaiNames))

	for i, oaiInfo := range oaiNames {
		oaiFile, err := gg.fs.Open(oaiInfo.File)
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

		ad, err := tyk_swagger.CreateDefinitionFromSwagger(&oai,
			oaiInfo.ORGID,
			oaiInfo.OAS.VersionName)
		if err != nil {
			return nil, err
		}

		if oaiInfo.APIID != "" {
			ad.APIID = oaiInfo.APIID
		}

		if oaiInfo.DBID != "" {
			ad.Id = bson.ObjectIdHex(oaiInfo.DBID)
		}

		if oaiInfo.OAS.OverrideListenPath != "" {
			ad.Proxy.ListenPath = oaiInfo.OAS.OverrideListenPath
		}

		if oaiInfo.OAS.OverrideTarget != "" {
			ad.Proxy.TargetURL = oaiInfo.OAS.OverrideTarget
		}

		if oaiInfo.OAS.StripListenPath {
			ad.Proxy.StripListenPath = true
		}

		defs[i] = *ad
	}

	return defs, nil
}

func (gg *GitGetter) FetchPolicies(spec *TykSourceSpec) ([]objects.Policy, error) {
	if gg.r == nil {
		return nil, errors.New("No repository in memory, fetch repo first")
	}

	defNames := spec.Policies
	defs := make([]objects.Policy, len(defNames))
	for i, defInfo := range defNames {
		defFile, err := gg.fs.Open(defInfo.File)
		if err != nil {
			fmt.Println(defInfo.File)
			return nil, err
		}

		rawDef, err := ioutil.ReadAll(defFile)
		if err != nil {
			return nil, err
		}

		pol := objects.Policy{}
		err = json.Unmarshal(rawDef, &pol)
		if err != nil {
			return nil, err
		}

		if defInfo.ID != "" {
			pol.ID = defInfo.ID
		}

		if pol.OrgID == "" {
			return nil, errors.New("Policies must include an org ID")
		}

		defs[i] = pol
	}

	fmt.Printf("Fetched %v policies\n", len(defs))

	return defs, nil
}

func (gg *GitGetter) Sync(apiDefs []apidef.APIDefinition) error {
	return gg.publisher.Sync(apiDefs)
}

func (gg *GitGetter) Create(apiDef *apidef.APIDefinition) (string, error) {
	return gg.publisher.Create(apiDef)
}

func (gg *GitGetter) Update(apiDef *apidef.APIDefinition) error {
	return gg.publisher.Update(apiDef)
}

func (gg *GitGetter) Reload() error {
	return gg.publisher.Reload()
}

func (gg *GitGetter) CreatePolicy(pol *objects.Policy) (string, error) {
	return gg.publisher.CreatePolicy(pol)
}

func (gg *GitGetter) UpdatePolicy(pol *objects.Policy) error {
	return gg.publisher.UpdatePolicy(pol)
}

func (gg *GitGetter) SyncPolicies(pols []objects.Policy) error {
	return gg.publisher.SyncPolicies(pols)
}
