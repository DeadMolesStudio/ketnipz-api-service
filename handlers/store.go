package handlers

import (
	"api/models"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/middleware"

	"api/database"
)

func SkinHandler(dm *db.DatabaseManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getSkin(w, r, dm)
		case http.MethodPost:
			buySkin(w, r, dm)
		case http.MethodPut:
			changeSkin(w, r, dm)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// @Summary Получить информацию об одном скине или обо всех
// @Description Получить информацию о скине: ID, название и стоимость
// @ID get-skin
// @Produce json
// @Param id query uint false "ID"
// @Success 200 {object} models.AllSkins "Скин/скины найдены"
// @Failure 400 "Неправильный запрос"
// @Failure 404 "Не найдено"
// @Failure 500 "Ошибка в бд"
// @Router /profile/skin [GET]
func getSkin(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager) {
	query := r.URL.Query()
	rawID := query.Get("id")
	var id uint64
	var err error
	if rawID != "" {
		id, err = strconv.ParseUint(rawID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if id != 0 {
		skin, err := database.GetSkin(dm, uint(id))
		if err != nil {
			if err == database.ErrNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			logger.Errorf("database error while getting skin with id %v: %v", id, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json, err := skin.MarshalJSON()
		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, string(json))
	} else {
		skins, err := database.GetAllSkins(dm)
		if err != nil {
			logger.Errorf("database error while getting all skins: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		skinsList := &models.AllSkins{
			Skins: *skins,
		}
		json, err := skinsList.MarshalJSON()
		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, string(json))
	}
}

// @Summary Купить новый скин
// @Description Купить новый скин, монет должно быть достаточно для совершения покупки
// @ID post-skin
// @Accept json
// @Param Profile body models.RequestSkin true "Скин для покупки"
// @Success 200 "Скин куплен (или уже есть)"
// @Failure 400 "Неверный формат JSON"
// @Failure 401 "Не залогинен, профиль не существует"
// @Failure 404 "Скин не найден"
// @Failure 422 "Недостаточно средств"
// @Failure 500 "Ошибка в бд"
// @Router /profile/skin [POST]
func buySkin(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager) {
	if !r.Context().Value(middleware.KeyIsAuthenticated).(bool) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	skin := &models.RequestSkin{}
	err := unmarshalJSONBodyToStruct(r, skin)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uID := r.Context().Value(middleware.KeyUserID).(uint)
	store, err := database.GetUserStore(dm, uID)
	if err != nil {
		switch err.(type) {
		case database.UserNotFoundError:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			logger.Errorf("database error while getting user store with id %v: %v", uID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	for _, v := range store.PurchasedSkins {
		if skin.ID == v {
			// user already has this skin
			return
		}
	}

	skinInfo, err := database.GetSkin(dm, skin.ID)
	if err != nil {
		if err == database.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		logger.Errorf("database error while getting skin with id %v: %v", skin.ID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if store.Coins != nil && *store.Coins < skinInfo.Cost {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = database.BuySkin(dm, uID, skinInfo)
	if err != nil {
		logger.Errorf("database error while buying skin %v by user %v: %v", *skinInfo, uID, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// @Summary Изменить скин
// @Description Изменить скин, должен быть куплен
// @ID put-skin
// @Accept json
// @Param Profile body models.RequestSkin true "Скин, который надеваем"
// @Success 200 "Пользователь найден, успешно надет скин, уже надет такой скин"
// @Failure 400 "Неверный формат JSON"
// @Failure 401 "Не залогинен, пользователь не существует"
// @Failure 422 "Скин не куплен"
// @Failure 500 "Ошибка в бд"
// @Router /profile/skin [PUT]
func changeSkin(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager) {
	if !r.Context().Value(middleware.KeyIsAuthenticated).(bool) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	skin := &models.RequestSkin{}
	err := unmarshalJSONBodyToStruct(r, skin)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uID := r.Context().Value(middleware.KeyUserID).(uint)
	store, err := database.GetUserStore(dm, uID)
	if err != nil {
		switch err.(type) {
		case database.UserNotFoundError:
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			logger.Errorf("database error while getting user store with id %v: %v", uID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if store.Skin != nil && *store.Skin == skin.ID {
		// equipped
		return
	}

	hasSkin := false
	if skin.ID != 0 {
		for _, v := range store.PurchasedSkins {
			if skin.ID == v {
				hasSkin = true
			}
		}
	} else {
		hasSkin = true // all users have default skin
	}

	if hasSkin {
		err = database.ChangeSkin(dm, uID, skin.ID)
		if err != nil {
			logger.Errorf("database error while changing user %v skin to %v: %v", uID, skin.ID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(http.StatusUnprocessableEntity)
}
