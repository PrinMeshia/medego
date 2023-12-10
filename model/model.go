package model

import up "github.com/upper/db/v4"

type Model interface {
	Table() string
}

func GetAllRecords(model Model, condition up.Cond, collection up.Collection) ([]interface{}, error) {
	res := collection.Find(condition)
	var all []interface{}

	if err := res.All(&all); err != nil {
		return nil, err
	}

	return all, nil
}
