package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func unmarshalJSONBodyToStruct(r *http.Request, s json.Unmarshaler) error {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	err = s.UnmarshalJSON(body)
	if err != nil {
		return ParseJSONError{err}
	}

	return nil
}
