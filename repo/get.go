package repo

import (
	"encoding/json"
	"fmt"

	"github.com/ghnexpress/traefik-cache/model"
)

func (r *repoManager) Get(key string) (*model.Cache, error) {
	b, err := r.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("Get data from redis error: %v", err)
	}

	if len(b) == 0 {
		return nil, nil
	}
	var d model.Cache
	if err = json.Unmarshal(b, &d); err != nil {
		return nil, fmt.Errorf("Unmarshal cache data error: %v", err)
	}

	return &d, nil
}
