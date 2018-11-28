package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"

	"api/models"
)

//easyjson:json
type ParseJSONError struct {
	msg error
}

func (e ParseJSONError) Error() string {
	return fmt.Sprintf("error while parsing JSON: %v", e.msg)
}

func sendError(w http.ResponseWriter, r *http.Request, e error, status int) {
	errStruct := models.Error{What: e.Error()}
	m, err := errStruct.MarshalJSON()
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintln(w, string(m))
}
