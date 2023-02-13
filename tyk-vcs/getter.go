package tyk_vcs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/TykTechnologies/tyk/apidef"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	tyk_swagger "github.com/TykTechnologies/tyk-sync/tyk-swagger"
)

type Getter interface {
	FetchRepo() error
	FetchAPIDef(spec *TykSourceSpec) ([]objects.DBApiDefinition, error)
	FetchPolicies(spec *TykSourceSpec) ([]objects.Policy, error)
	FetchTykSpec() (*TykSourceSpec, error)
}

type BaseGetter struct {
	Getter
	fs billy.Filesystem
}

type GitGetter struct {
	*BaseGetter
	Getter
	repo             string
	branch           string
	key              []byte
	fs               billy.Filesystem
	r                *git.Repository
	subdirectoryPath string
}

type FSGetter struct {
	*BaseGetter
	Getter
	fs               billy.Filesystem
	subdirectoryPath string
}

func NewGGetter(repo, branch string, key []byte, subdirectoryPath string) (*GitGetter, error) {
	gh := &GitGetter{
		repo:             repo,
		branch:           branch,
		key:              key,
		fs:               memfs.New(),
		subdirectoryPath: subdirectoryPath,
	}

	return gh, nil
}

func NewFSGetter(filePath string, subdirectoryPath string) (*FSGetter, error) {
	gh := &FSGetter{
		fs:               osfs.New(filePath),
		subdirectoryPath: subdirectoryPath,
	}

	return gh, nil
}

func (gg *GitGetter) FetchRepo() error {

	cloneOptions := git.CloneOptions{
		URL:           gg.repo,
		ReferenceName: plumbing.ReferenceName(gg.branch),
		SingleBranch:  true,
	}
	if len(gg.key) != 0 {
		publicKey, keyError := ssh.NewPublicKeys("git", gg.key, "")
		if keyError != nil {
			fmt.Println("Error getting key for git authentication:", keyError)
		}
		cloneOptions.Auth = publicKey
	}
	r, err := git.Clone(memory.NewStorage(), gg.fs, &cloneOptions)

	if err != nil {
		return err
	}

	gg.r = r

	return nil
}

func (gg *FSGetter) FetchRepo() error {
	return nil
}

func fetchSpec(fs billy.Filesystem, subdirectoryPath string) (*TykSourceSpec, error) {
	specFile, err := fs.Open(getFilepath(".tyk.json", subdirectoryPath))
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
	return fetchSpec(gg.fs, gg.subdirectoryPath)
}

func (gg *FSGetter) FetchTykSpec() (*TykSourceSpec, error) {
	return fetchSpec(gg.fs, gg.subdirectoryPath)
}

func (gg *FSGetter) FetchAPIDef(spec *TykSourceSpec) ([]objects.DBApiDefinition, error) {
	return fetchAPIDefinitions(gg.fs, spec, gg.subdirectoryPath)
}

func (gg *GitGetter) FetchAPIDef(spec *TykSourceSpec) ([]objects.DBApiDefinition, error) {
	if gg.r == nil {
		return nil, errors.New("no repository in memory, fetch repo first")
	}
	return fetchAPIDefinitions(gg.fs, spec, gg.subdirectoryPath)
}

func fetchAPIDefinitions(fs billy.Filesystem, spec *TykSourceSpec, subdirectoryPath string) ([]objects.DBApiDefinition, error) {
	switch spec.Type {
	case TYPE_APIDEF:
		return fetchAPIDefinitionsDirect(fs, spec, subdirectoryPath)
	case TYPE_OAI:
		return fetchAPIDefinitionsFromOAI(fs, spec, subdirectoryPath)
	default:
		return nil, fmt.Errorf("Type must be '%v or '%v'", TYPE_APIDEF, TYPE_OAI)
	}
}

func fetchAPIDefinitionsDirect(fs billy.Filesystem, spec *TykSourceSpec, subdirectoryPath string) ([]objects.DBApiDefinition, error) {
	defNames := spec.Files
	defs := make([]objects.DBApiDefinition, len(defNames))
	for i, defInfo := range defNames {
		defFile, err := fs.Open(getFilepath(defInfo.File, subdirectoryPath))
		if err != nil {
			return nil, err
		}

		rawDef, err := ioutil.ReadAll(defFile)
		if err != nil {
			return nil, err
		}

		ad := objects.DBApiDefinition{}
		err = json.Unmarshal(rawDef, &ad)
		if err != nil || (ad.APIDefinition == nil) {
			def := objects.APIDefinition{}
			errSecondUnmarshal := json.Unmarshal(rawDef, &def)
			if errSecondUnmarshal != nil {
				return nil, err
			}
			ad.APIDefinition = &def
		}

		if defInfo.APIID != "" {
			ad.APIID = defInfo.APIID
		}

		if defInfo.DBID != "" {
			ad.Id = apidef.ObjectIdHex(defInfo.DBID)
		}

		if defInfo.ORGID != "" {
			ad.OrgID = defInfo.ORGID
		}

		defs[i] = ad
	}

	fmt.Printf("Fetched %v definitions\n", len(defs))
	return defs, nil
}

func fetchAPIDefinitionsFromOAI(fs billy.Filesystem, spec *TykSourceSpec, subdirectoryPath string) ([]objects.DBApiDefinition, error) {
	oaiNames := spec.Files
	defs := make([]objects.DBApiDefinition, len(oaiNames))

	for i, oaiInfo := range oaiNames {
		oaiFile, err := fs.Open(getFilepath(oaiInfo.File, subdirectoryPath))
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
			ad.Id = apidef.ObjectIdHex(oaiInfo.DBID)
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
	return fetchPolicies(gg.fs, spec, gg.subdirectoryPath)
}

func (gg *GitGetter) FetchPolicies(spec *TykSourceSpec) ([]objects.Policy, error) {
	if gg.r == nil {
		return nil, errors.New("No repository in memory, fetch repo first")
	}
	return fetchPolicies(gg.fs, spec, gg.subdirectoryPath)
}

func fetchPolicies(fs billy.Filesystem, spec *TykSourceSpec, subdirectoryPath string) ([]objects.Policy, error) {
	defNames := spec.Policies
	defs := make([]objects.Policy, len(defNames))
	for i, defInfo := range defNames {
		defFile, err := fs.Open(getFilepath(defInfo.File, subdirectoryPath))
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

func getFilepath(file string, pathSegments ...string) string {
	if len(pathSegments) == 0 {
		return file
	}
	allSegments := append(pathSegments, file)
	return filepath.Join(allSegments...)
}
