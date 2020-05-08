package tyk_vcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk-sync/tyk-swagger"
	"github.com/TykTechnologies/tyk/apidef"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"io/ioutil"
)

type Getter interface {
	FetchRepo() error
	FetchAPIDef(spec *TykSourceSpec) ([]apidef.APIDefinition, error)
	FetchPolicies(spec *TykSourceSpec) ([]objects.Policy, error)
	FetchTykSpec() (*TykSourceSpec, error)
}

type BaseGetter struct {
	Getter
	fs        billy.Filesystem
}

type GitGetter struct {
	*BaseGetter
	Getter
	repo      string
	branch    string
	key       []byte
	fs        billy.Filesystem
	r         *git.Repository
}

type FSGetter struct {
	*BaseGetter
	Getter
	fs        billy.Filesystem
}

func NewGGetter(repo, branch string, key []byte) (*GitGetter, error) {
	gh := &GitGetter{
		repo:      repo,
		branch:    branch,
		key:       key,
		fs:        memfs.New(),
	}

	return gh, nil
}

func NewFSGetter(filePath string) (*FSGetter, error) {
	gh := &FSGetter{
		fs:        osfs.New(filePath),
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

func (gg *FSGetter) FetchRepo() error {
	return nil
}

func fetchSpec(fs billy.Filesystem) (*TykSourceSpec, error) {
	specFile, err := fs.Open(".tyk.json")
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

func (gg *GitGetter) FetchTykSpec() (*TykSourceSpec, error) {
	if gg.r == nil {
		return nil, errors.New("no repository in memory, fetch repo first")
	}
	return fetchSpec(gg.fs)
}

func (gg *FSGetter) FetchTykSpec() (*TykSourceSpec, error) {
	return fetchSpec(gg.fs)
}

func (gg *FSGetter) FetchAPIDef(spec *TykSourceSpec) ([]apidef.APIDefinition, error) {
	return fetchAPIDefinitions(gg.fs, spec)
}

func (gg *GitGetter) FetchAPIDef(spec *TykSourceSpec) ([]apidef.APIDefinition, error) {
	if gg.r == nil {
		return nil, errors.New("no repository in memory, fetch repo first")
	}
	return fetchAPIDefinitions(gg.fs, spec)
}

func fetchAPIDefinitions(fs billy.Filesystem, spec *TykSourceSpec) ([]apidef.APIDefinition, error) {
	switch spec.Type {
	case TYPE_APIDEF:
		return fetchAPIDefinitionsDirect(fs, spec)
	case TYPE_OAI:
		return fetchAPIDefinitionsFromOAI(fs, spec)
	default:
		return nil, fmt.Errorf("Type must be '%v or '%v'", TYPE_APIDEF, TYPE_OAI)
	}
}

func fetchAPIDefinitionsDirect(fs billy.Filesystem, spec *TykSourceSpec) ([]apidef.APIDefinition, error) {
	defNames := spec.Files
	defs := make([]apidef.APIDefinition, len(defNames))
	for i, defInfo := range defNames {
		defFile, err := fs.Open(defInfo.File)
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

func fetchAPIDefinitionsFromOAI(fs billy.Filesystem, spec *TykSourceSpec) ([]apidef.APIDefinition, error) {
	oaiNames := spec.Files
	defs := make([]apidef.APIDefinition, len(oaiNames))

	for i, oaiInfo := range oaiNames {
		oaiFile, err := fs.Open(oaiInfo.File)
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

func (gg *FSGetter) FetchPolicies(spec *TykSourceSpec) ([]objects.Policy, error) {
	return fetchPolicies(gg.fs, spec)
}

func (gg *GitGetter) FetchPolicies(spec *TykSourceSpec) ([]objects.Policy, error) {
	if gg.r == nil {
		return nil, errors.New("No repository in memory, fetch repo first")
	}
	return fetchPolicies(gg.fs, spec)
}

func fetchPolicies(fs billy.Filesystem, spec *TykSourceSpec) ([]objects.Policy, error)  {
	defNames := spec.Policies
	defs := make([]objects.Policy, len(defNames))
	for i, defInfo := range defNames {
		defFile, err := fs.Open(defInfo.File)
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
