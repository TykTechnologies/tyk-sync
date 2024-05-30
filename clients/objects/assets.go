package objects

import (
	"errors"
	"time"

	"github.com/TykTechnologies/storage/persistent/model"
)

type DBAssets struct {
	DBId        model.ObjectID `json:"_id"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	OrgID       string         `json:"org_id"`
	Description string         `json:"description"`
	Kind        string         `json:"kind"`
	Data        JSONRawMessage `bson:"data" json:"data"`
	LastUpdated time.Time      `json:"last_updated"`
}

// JSONRawMessage implements Scanner and Valuer interface for gorm.
type JSONRawMessage []byte

// MarshalJSON returns m as the JSON encoding of m.
func (j JSONRawMessage) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}

	return j, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (j *JSONRawMessage) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}

	*j = append((*j)[0:0], data...)

	return nil
}
