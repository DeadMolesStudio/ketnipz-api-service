package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/middleware"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/session"

	"api/database"
	"api/filesystem"
	"api/models"
)

func cleanProfile(r *http.Request, p *models.RegisterProfile) error {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	err = p.UnmarshalJSON(body)
	if err != nil {
		return ParseJSONError{err}
	}

	return nil
}

func validateNickname(dm *db.DatabaseManager, s string) ([]models.ProfileError, error) {
	var errors []models.ProfileError

	isValid := govalidator.StringLength(s, "4", "32")
	if !isValid {
		errors = append(errors, models.ProfileError{
			Field: "nickname",
			Text:  "Никнейм должен быть не менее 4 символов и не более 32 символов",
		})
		return errors, nil
	}

	exists, err := database.CheckExistenceOfNickname(dm, s)
	if err != nil {
		logger.Error(err)
		return errors, err
	}
	if exists {
		errors = append(errors, models.ProfileError{
			Field: "nickname",
			Text:  "Этот никнейм уже занят",
		})
	}

	return errors, nil
}

func validateEmail(dm *db.DatabaseManager, s string) ([]models.ProfileError, error) {
	var errors []models.ProfileError

	isValid := govalidator.IsEmail(s)
	if !isValid {
		errors = append(errors, models.ProfileError{
			Field: "email",
			Text:  "Невалидная почта",
		})
		return errors, nil
	}

	exists, err := database.CheckExistenceOfEmail(dm, s)
	if err != nil {
		logger.Error(err)
		return errors, err
	}
	if exists {
		errors = append(errors, models.ProfileError{
			Field: "email",
			Text:  "Данная почта уже занята",
		})
	}

	return errors, nil
}

func validatePassword(s string) []models.ProfileError {
	var errors []models.ProfileError

	isValid := govalidator.StringLength(s, "8", "32")
	if !isValid {
		errors = append(errors, models.ProfileError{
			Field: "password",
			Text:  "Пароль должен быть не менее 8 символов и не более 32 символов",
		})
	}

	return errors
}

func validateFields(dm *db.DatabaseManager, u *models.RegisterProfile) ([]models.ProfileError, error) {
	var errors []models.ProfileError

	valErrors, dbErr := validateNickname(dm, u.Nickname)
	if dbErr != nil {
		return []models.ProfileError{}, dbErr
	}
	errors = append(errors, valErrors...)

	valErrors, dbErr = validateEmail(dm, u.Email)
	if dbErr != nil {
		return []models.ProfileError{}, dbErr
	}
	errors = append(errors, valErrors...)
	errors = append(errors, validatePassword(u.Password)...)

	return errors, nil
}

func ProfileHandler(dm *db.DatabaseManager, sm *session.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getProfile(w, r, dm)
		case http.MethodPost:
			postProfile(w, r, dm, sm)
		case http.MethodPut:
			putProfile(w, r, dm)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// @Summary Получить профиль
// @Description Получить профиль пользователя по ID, никнейму или из сессии
// @ID get-profile
// @Produce json
// @Param id query uint false "ID"
// @Param nickname query string false "Никнейм"
// @Success 200 {object} models.Profile "Пользователь найден, успешно"
// @Failure 400 "Неправильный запрос"
// @Failure 401 "Не залогинен"
// @Failure 404 "Не найдено"
// @Failure 500 "Ошибка в бд"
// @Router /profile [GET]
func getProfile(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager) {
	params := &models.RequestProfile{}
	err := decoder.Decode(params, r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if params.ID != 0 {
		profile, err := database.GetUserProfileByID(dm, params.ID)
		if err != nil {
			switch err.(type) {
			case database.UserNotFoundError:
				w.WriteHeader(http.StatusNotFound)
				return
			default:
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json, err := profile.MarshalJSON()
		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, string(json))
	} else if params.Nickname != "" {
		profile, err := database.GetUserProfileByNickname(dm, params.Nickname)
		if err != nil {
			switch err.(type) {
			case database.UserNotFoundError:
				w.WriteHeader(http.StatusNotFound)
				return
			default:
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json, err := profile.MarshalJSON()
		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, string(json))
	} else {
		if !r.Context().Value(middleware.KeyIsAuthenticated).(bool) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		profile, err := database.GetUserProfileByID(dm, r.Context().Value(middleware.KeyUserID).(uint))
		if err != nil {
			switch err.(type) {
			case database.UserNotFoundError:
				w.WriteHeader(http.StatusNotFound)
				return
			default:
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json, err := profile.MarshalJSON()
		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, string(json))
	}
}

// @Summary Зарегистрироваться и залогиниться по новому профилю
// @Description Зарегистрировать по никнейму, почте и паролю и автоматически залогинить
// @ID post-profile
// @Accept json
// @Produce json
// @Param Profile body models.RegisterProfile true "Никнейм, почта и пароль"
// @Success 200 "Пользователь зарегистрирован и залогинен успешно"
// @Failure 400 "Неверный формат JSON"
// @Failure 403 {object} models.ProfileErrorList "Ошибки при регистрации: невалидна или занята почта, занят ник, пароль не удовлетворяет правилам безопасности, другие ошибки"
// @Failure 422 "При регистрации не все параметры"
// @Failure 500 "Ошибка в бд"
// @Router /profile [POST]
func postProfile(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager, sm *session.SessionManager) {
	u := &models.RegisterProfile{}
	err := cleanProfile(r, u)
	if err != nil {
		switch err.(type) {
		case ParseJSONError:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if u.Nickname == "" || u.Email == "" || u.Password == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	fieldErrors, err := validateFields(dm, u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(fieldErrors) != 0 {
		sendList := models.ProfileErrorList{Errors: fieldErrors}
		json, err := sendList.MarshalJSON()
		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, string(json))
	} else {
		newU, err := database.CreateNewUser(dm, u)
		if err != nil {
			if err == db.ErrUniqueConstraintViolation ||
				err == db.ErrNotNullConstraintViolation {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = loginUser(w, sm, newU.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Infof("New user with id %v, email %v and nickname %v logged in", newU.UserID, newU.Email, newU.Nickname)
	}
}

// @Summary Изменить профиль
// @Description Изменить профиль, должен быть залогинен
// @ID put-profile
// @Accept json
// @Produce json
// @Param Profile body models.RegisterProfile true "Новые никнейм, и/или почта, и/или пароль"
// @Success 200 "Пользователь найден, успешно изменены данные"
// @Failure 400 "Неверный формат JSON"
// @Failure 401 "Не залогинен"
// @Failure 403 {object} models.ProfileErrorList "Ошибки при регистрации: невалидна или занята почта, занят ник, пароль не удовлетворяет правилам безопасности, другие ошибки"
// @Failure 500 "Ошибка в бд"
// @Router /profile [PUT]
func putProfile(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager) {
	if !r.Context().Value(middleware.KeyIsAuthenticated).(bool) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u := &models.RegisterProfile{}
	err := cleanProfile(r, u)
	if err != nil {
		switch err.(type) {
		case ParseJSONError:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if u.Nickname == "" && u.Email == "" && u.Password == "" {
		return
	}

	var fieldErrors []models.ProfileError

	if u.Nickname != "" {
		valErrors, dbErr := validateNickname(dm, u.Nickname)
		if dbErr != nil {
			logger.Error(dbErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fieldErrors = append(fieldErrors, valErrors...)
	}
	if u.Email != "" {
		valErrors, dbErr := validateEmail(dm, u.Email)
		if dbErr != nil {
			logger.Error(dbErr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fieldErrors = append(fieldErrors, valErrors...)
	}
	if u.Password != "" {
		fieldErrors = append(fieldErrors, validatePassword(u.Password)...)
	}

	if len(fieldErrors) != 0 {
		sendList := models.ProfileErrorList{Errors: fieldErrors}
		json, err := sendList.MarshalJSON()
		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, string(json))
	} else {
		id := r.Context().Value(middleware.KeyUserID).(uint)
		err := database.UpdateUserByID(dm, id, u)
		if err != nil {
			switch err.(type) {
			case database.UserNotFoundError:
				w.WriteHeader(http.StatusNotFound)
			default:
				logger.Error(err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		logger.Infof("user with id %v changed to %v %v", id, u.Nickname, u.Email)
	}
}

func AvatarHandler(dm *db.DatabaseManager) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			putAvatar(w, r, dm)
		case http.MethodDelete:
			deleteAvatar(w, r, dm)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// @Summary Изменить аватар
// @Description Загрузить или изменить уже существующий аватар
// @ID put-avatar
// @Accept multipart/form-data
// @Success 200 "Удалена аватарка у пользователя"
// @Failure 401 "Не залогинен"
// @Failure 404 "Пользователь не найден"
// @Failure 500 "Ошибка при парсинге, в бд, файловой системе"
// @Router /profile/avatar [PUT]
func putAvatar(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager) {
	if !r.Context().Value(middleware.KeyIsAuthenticated).(bool) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(5 * (1 << 20)) // 5 MB
	if err != nil {
		if err == http.ErrNotMultipart || err == http.ErrMissingBoundary {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	avatar, fileHeader, err := r.FormFile("avatar")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer avatar.Close()

	uID := r.Context().Value(middleware.KeyUserID).(uint)
	filename := fileHeader.Filename
	dir := "static/img/"
	filename = filesystem.GetHashedNameForFile(uID, filename)
	err = filesystem.SaveFile(avatar, dir, filename)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = database.UploadAvatar(dm, uID, "/"+dir+filename)
	if err != nil {
		switch err.(type) {
		case *database.UserNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}

// @Summary Удалить аватар
// @Description Удалить аватар, пользователь должен быть залогинен
// @ID delete-avatar
// @Success 200 "Удалена аватарка у пользователя"
// @Failure 401 "Не залогинен"
// @Failure 404 "Пользователь не найден"
// @Failure 500 "Ошибка в бд"
// @Router /profile/avatar [DELETE]
func deleteAvatar(w http.ResponseWriter, r *http.Request, dm *db.DatabaseManager) {
	if !r.Context().Value(middleware.KeyIsAuthenticated).(bool) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := database.DeleteAvatar(dm, r.Context().Value(middleware.KeyUserID).(uint))
	if err != nil {
		switch err.(type) {
		case *database.UserNotFoundError:
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}
