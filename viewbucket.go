package main

import (
	"errors"

	_ "github.com/mattes/migrate/driver/postgres"
) //for migrations

type viewBucket struct {
	Data map[string]interface{}
}

func newViewBucket() *viewBucket {
	return &viewBucket{Data: map[string]interface{}{}}
}

func (viewBucket *viewBucket) add(key string, value interface{}) {
	viewBucket.Data[key] = value
}

func (viewBucket *viewBucket) getStrSafe(key string) (string, error) {
	val := viewBucket.Data[key]
	if val == nil {
		return "", errors.New("could not find " + key)
	}
	strVal, ok := val.(string)
	if !ok {
		return "", errors.New("could not cast " + key + " to string")
	}
	return strVal, nil
}

// getStr - returns a string for the provided key. Will panic if key not found
func (viewBucket *viewBucket) getStr(key string) string {
	val, err := viewBucket.getStrSafe(key)
	if err != nil {
		panic(err)
	}
	return val
}
