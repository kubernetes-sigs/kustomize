package doc

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"google.golang.org/appengine/search"
)

const (
	identifierStr   = "identifier"
	documentStr     = "document"
	repoURLStr      = "repo_url"
	filePathStr     = "file_path"
	creationTimeStr = "creation_time"
)

// Represents an unbreakable character stream.
type Atom = search.Atom

// Implements search.FieldLoadSaver in order to index this representation of a kustomization.yaml
// file.
type KustomizationDocument struct {
	identifiers   []Atom
	FilePath      Atom
	RepositoryURL Atom
	DocumentData  string
	CreationTime  time.Time
}

// Partially implements search.FieldLoadSaver.
func (k *KustomizationDocument) Load(fields []search.Field, metadata *search.DocumentMetadata) error {
	k.identifiers = make([]search.Atom, 0)
	wrongTypeError := func(name string, expected interface{}, actual interface{}) error {
		return fmt.Errorf("%s expects type %T, found %#v", name, expected, actual)
	}

	for _, f := range fields {
		switch f.Name {
		case identifierStr:
			identifier, ok := f.Value.(search.Atom)
			if !ok {
				return wrongTypeError(f.Name, identifier, f.Value)
			}
			k.identifiers = append(k.identifiers, identifier)

		case documentStr:
			document, ok := f.Value.(string)
			if !ok {
				return wrongTypeError(f.Name, document, f.Value)
			}
			k.DocumentData = document

		case filePathStr:
			fp, ok := f.Value.(search.Atom)
			if !ok {
				return wrongTypeError(f.Name, fp, f.Value)
			}
			k.FilePath = fp

		case repoURLStr:
			url, ok := f.Value.(search.Atom)
			if !ok {
				return wrongTypeError(f.Name, url, f.Value)
			}
			k.RepositoryURL = url

		case creationTimeStr:
			time, ok := f.Value.(time.Time)
			if !ok {
				return wrongTypeError(f.Name, time, f.Value)
			}
			k.CreationTime = time
		default:
			return fmt.Errorf("KustomizationDocument field %s not recognized", f.Name)
		}
	}

	return nil
}

// Partially implements search.FieldLoadSaver.
func (k *KustomizationDocument) Save() ([]search.Field, *search.DocumentMetadata, error) {
	err := k.ParseYAML()
	if err != nil {
		return nil, nil, err
	}

	extraFields := []search.Field{
		{Name: documentStr, Value: k.DocumentData},
		{Name: filePathStr, Value: k.FilePath},
		{Name: repoURLStr, Value: k.RepositoryURL},
		{Name: creationTimeStr, Value: k.CreationTime},
	}

	fields := make([]search.Field, 0, len(k.identifiers)+len(extraFields))
	for _, identifier := range k.identifiers {
		fields = append(fields, search.Field{Name: identifierStr, Value: identifier})
	}
	fields = append(fields, extraFields...)

	return fields, nil, nil
}

func (k *KustomizationDocument) ParseYAML() error {
	k.identifiers = make([]Atom, 0)

	var kustomization map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(k.DocumentData), &kustomization)
	if err != nil {
		return fmt.Errorf("unable to parse kustomization file: %s", err)
	}

	type Map struct {
		data   map[interface{}]interface{}
		prefix Atom
	}

	toVisit := []Map{
		{
			data:   kustomization,
			prefix: "",
		},
	}

	atomJoin := func(vals ...interface{}) Atom {
		strs := make([]string, 0, len(vals))
		for _, val := range vals {
			strs = append(strs, fmt.Sprint(val))
		}
		return Atom(strings.Trim(strings.Join(strs, " "), " "))
	}

	set := make(map[Atom]struct{})

	for i := 0; i < len(toVisit); i++ {
		visiting := toVisit[i]
		for k, v := range visiting.data {
			set[atomJoin(visiting.prefix, k)] = struct{}{}
			switch value := v.(type) {
			case map[interface{}]interface{}:
				toVisit = append(toVisit, Map{
					data:   value,
					prefix: atomJoin(visiting.prefix, fmt.Sprint(k)),
				})
			case []interface{}:
				for _, val := range value {
					submap, ok := val.(map[interface{}]interface{})
					if !ok {
						continue
					}
					toVisit = append(toVisit, Map{
						data:   submap,
						prefix: atomJoin(visiting.prefix, fmt.Sprint(k)),
					})
				}
			}
		}
	}

	for key := range set {
		k.identifiers = append(k.identifiers, key)
	}

	return nil
}
