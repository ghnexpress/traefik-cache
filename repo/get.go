package repo

import (
	"encoding/json"
	"fmt"

	"github.com/ghnexpress/traefik-cache/model"
)

func (r *repoManager) Get(key string) (data *model.Cache, err error) {
	b, err := r.db.Get(key)
	if err != nil {
		err = fmt.Errorf("Get data from redis error: %v", err)
		return
	}

	if len(b) == 0 {
		return nil, nil
	}
	var d model.Cache
	if err = json.Unmarshal(b, &d); err != nil {
		err = fmt.Errorf("Unmarshal cache data error: %v", err)
		return
	}

	return
}
