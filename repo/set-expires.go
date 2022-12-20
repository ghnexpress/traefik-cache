package repo

import (
	"encoding/json"
	"fmt"

	"github.com/ghnexpress/traefik-cache/model"
)

func (r *repoManager) SetExpires(key string, time int64, data model.Cache) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Marshal cache data error: %v", err)
	}

	if err = r.db.Setex(key, time, b); err != nil {
		return fmt.Errorf("Set data to redis error: %v", err)
	}

	// if err = r.db.Set(key, b); err != nil {
	// 	return fmt.Errorf("Set data to redis error: %v", err)
	// }

	// var ok bool
	// if ok, err = r.db.Expire(key, time); err != nil {
	// 	return fmt.Errorf("Expire data to redis error: %v", err)
	// }

	return nil
}
