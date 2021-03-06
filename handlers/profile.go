package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"golang.org/x/crypto/bcrypt"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/logger"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/middleware"
	"github.com/go-park-mail-ru/2018_2_DeadMolesStudio/session"

	"api/database"
	"api/filesystem"
	"api/models"
)

func validateNickname(dm *db.DatabaseManager, s string) ([]models.ProfileError, error) {
	var errors []models.ProfileError

	isValid := govalidator.StringLength(s, "4", "20")
	if !isValid {
		errors = append(errors, models.ProfileError{
			Field: "nickname",
			Text:  "Nickname must be at least 4 characters and no more than 20 characters.",
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
			Text:  "This nickname is already taken.",
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
			Text:  "Invalid email.",
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
			Text:  "This email is already taken.",
		})
	}

	return errors, nil
}

func validatePassword(s string) []models.ProfileError {
	var errors []models.ProfileError

	isValid := govalidator.StringLength(s, "4", "32")
	if !isValid {
		errors = append(errors, models.ProfileError{
			Field: "password",
			Text:  "Password must be at least 4 characters and no more than 32 characters.",
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

func hashAndSalt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), err
}

func comparePasswords(hashed string, clean string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(clean))
	switch err {
	case nil:
		return true, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return false, nil
	default:
		return false, err
	}
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
		profile, err := database.GetUserProfileByID(dm, uint(id), false)
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
		return
	}
	nickname := query.Get("nickname")
	if nickname != "" {
		profile, err := database.GetUserProfileByNickname(dm, nickname)
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
		return
	}

	if !r.Context().Value(middleware.KeyIsAuthenticated).(bool) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	profile, err := database.GetUserProfileByID(dm, r.Context().Value(middleware.KeyUserID).(uint), true)
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
	err := unmarshalJSONBodyToStruct(r, u)
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
		u.Password, err = hashAndSalt(u.Password)
		if err != nil {
			logger.Errorf("hash and salt password error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	err := unmarshalJSONBodyToStruct(r, u)
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
	return func(w http.ResponseWriter, r *http.Request) {
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
	filename, err = filesystem.GetHashedNameForFile(uID, filename)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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

func CheckAvailabilityHandler(dm *db.DatabaseManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			query := r.URL.Query()
			nickname := query.Get("nickname")
			if nickname != "" {
				exists, err := database.CheckExistenceOfNickname(dm, nickname)
				if err != nil {
					logger.Errorf("check availability error: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if exists {
					w.WriteHeader(http.StatusForbidden)
				}
				return
			}
			email := query.Get("email")
			if email != "" {
				exists, err := database.CheckExistenceOfEmail(dm, email)
				if err != nil {
					logger.Errorf("check availability error: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				if exists {
					w.WriteHeader(http.StatusForbidden)
				}
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
