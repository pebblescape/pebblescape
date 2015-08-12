package httphelper

import (
	"encoding/json"
	"net/http"
	"reflect"
)

func JSON(w http.ResponseWriter, status int, v interface{}) {
	// Encode nil slices as `[]` instead of `null`
	if rv := reflect.ValueOf(v); rv.Type().Kind() == reflect.Slice && rv.IsNil() {
		v = []struct{}{}
	}

	var result []byte
	var err error
	result, err = json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(result)
}

func DecodeJSON(req *http.Request, i interface{}) error {
	dec := json.NewDecoder(req.Body)
	return dec.Decode(i)
}
