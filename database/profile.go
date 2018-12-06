package database

import (
	"database/sql"
	"strings"

	"github.com/lib/pq"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"

	"api/models"
)

func GetUserPassword(dm *db.DatabaseManager, e string) (*models.User, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}
	res := &models.User{}
	err = dbo.Get(res, `
	SELECT user_id, email, password FROM user_profile
	WHERE email = $1`,
		e)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, UserNotFoundError{"email"}
		}
		return res, err
	}

	return res, nil
}

func CreateNewUser(dm *db.DatabaseManager, u *models.RegisterProfile) (*models.Profile, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}
	qres := dbo.QueryRowx(`
		INSERT INTO user_profile (email, password, nickname)
		VALUES ($1, $2, $3) RETURNING user_id, email, nickname`,
		u.Email, u.Password, u.Nickname)
	if err := qres.Err(); err != nil {
		pqErr := err.(*pq.Error)
		switch pqErr.Code {
		case "23502":
			return nil, db.ErrNotNullConstraintViolation
		case "23505":
			return nil, db.ErrUniqueConstraintViolation
		}
	}
	res := &models.Profile{}
	err = qres.StructScan(res)
	if err != nil {
		return res, err
	}

	return res, nil
}

func UpdateUserByID(dm *db.DatabaseManager, id uint, u *models.RegisterProfile) error {
	if u.Email == "" && u.Password == "" && u.Nickname == "" {
		return nil
	}

	q := strings.Builder{}
	q.WriteString(`
		UPDATE user_profile
		SET `)
	hasBefore := false
	if u.Email != "" {
		q.WriteString("email = :email")
		hasBefore = true
	}
	if u.Password != "" {
		if hasBefore {
			q.WriteString(", ")
		}
		q.WriteString("password = :password")
		hasBefore = true
	}
	if u.Nickname != "" {
		if hasBefore {
			q.WriteString(", ")
		}
		q.WriteString("nickname = :nickname")
	}
	q.WriteString(`
		WHERE user_id = :user_id`)

	dbo, err := dm.DB()
	if err != nil {
		return err
	}
	_, err = dbo.NamedExec(q.String(), &models.Profile{
		User: models.User{
			UserID: id,
			UserPassword: models.UserPassword{
				Email:    u.Email,
				Password: u.Password,
			},
		},
		Nickname: u.Nickname,
	})
	if err != nil {
		return err
	}

	return nil
}

func GetUserProfileByID(dm *db.DatabaseManager, id uint, with_email bool) (*models.Profile, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}
	res := &models.Profile{}
	q := ""
	if with_email {
		q = `
		SELECT user_id, email, nickname, avatar, record, win, draws, loss FROM user_profile
		WHERE user_id = $1`
	} else {
		q = `
		SELECT user_id, nickname, avatar, record, win, draws, loss FROM user_profile
		WHERE user_id = $1`
	}
	err = dbo.Get(res, q, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, UserNotFoundError{"id"}
		}
		return res, err
	}

	return res, nil
}

func GetUserProfileByNickname(dm *db.DatabaseManager, nickname string) (*models.Profile, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}
	res := &models.Profile{}
	err = dbo.Get(res, `
		SELECT user_id, nickname, avatar, record, win, draws, loss FROM user_profile
		WHERE nickname = $1`,
		nickname)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, UserNotFoundError{"nickname"}
		}
		return res, err
	}

	return res, nil
}

func CheckExistenceOfEmail(dm *db.DatabaseManager, e string) (bool, error) {
	dbo, err := dm.DB()
	if err != nil {
		return false, err
	}
	res := &models.Profile{}
	err = dbo.Get(res, `
	SELECT FROM user_profile
	WHERE email = $1`,
		e)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func CheckExistenceOfNickname(dm *db.DatabaseManager, n string) (bool, error) {
	dbo, err := dm.DB()
	if err != nil {
		return false, err
	}
	res := &models.Profile{}
	err = dbo.Get(res, `
		SELECT FROM user_profile
		WHERE nickname = $1`,
		n)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func GetCountOfUsers(dm *db.DatabaseManager) (int, error) {
	dbo, err := dm.DB()
	if err != nil {
		return 0, err
	}
	// TODO: optimize it
	res := 0
	err = dbo.Get(&res, `
	SELECT COUNT(*) FROM user_profile`)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func UploadAvatar(dm *db.DatabaseManager, uID uint, path string) error {
	dbo, err := dm.DB()
	if err != nil {
		return err
	}
	qres, err := dbo.Exec(`
		UPDATE user_profile
		SET avatar = $2
		WHERE user_id = $1`,
		uID, path)
	if err != nil {
		return err
	}
	res, err := qres.RowsAffected()
	if err != nil {
		return err
	}
	if res == 0 {
		return &UserNotFoundError{"id"}
	}

	return nil
}

func DeleteAvatar(dm *db.DatabaseManager, uID uint) error {
	dbo, err := dm.DB()
	if err != nil {
		return err
	}
	qres, err := dbo.Exec(`
		UPDATE user_profile
		SET avatar = NULL
		WHERE user_id = $1`,
		uID)
	if err != nil {
		return err
	}
	res, err := qres.RowsAffected()
	if err != nil {
		return err
	}
	if res == 0 {
		return &UserNotFoundError{"id"}
	}

	return nil
}
