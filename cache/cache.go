package cache

import (
	"bytes"
	"encoding/gob"
)


func encode(item Entry) ([]byte, error) {
	var b bytes.Buffer
	e := gob.NewEncoder(&b)
	err := e.Encode(item)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func decode(data string) (Entry, error) {
	var item Entry
	b := bytes.Buffer{}
	b.Write([]byte(data))
	d := gob.NewDecoder(&b)
	err := d.Decode(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

