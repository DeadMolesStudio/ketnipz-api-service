package handlers

import (
	"fmt"
	"net/http"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"

	"api/database"
	"api/models"
)

// @Summary Получить таблицу лидеров
// @Description Получить таблицу лидеров (пагинация присутствует)
// @ID get-scoreboard
// @Produce json
// @Param Limit query uint false "Пользователей на страницу"
// @Param Page query uint false "Страница номер"
// @Success 200 {object} models.PositionList "Таблицу лидеров или ее страница и общее количество"
// @Failure 500 "Ошибка в бд"
// @Router /scoreboard [GET]
func ScoreboardHandler(dm *db.DatabaseManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			params := &models.FetchScoreboardPage{}
			err := decoder.Decode(params, r.URL.Query())
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			records, total, err := database.GetUserPositionsDescendingPaginated(
				dm, params)
			if err != nil {
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			positionsList := models.PositionList{
				List:  records,
				Total: total,
			}
			json, err := positionsList.MarshalJSON()
			if err != nil {
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, string(json))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
