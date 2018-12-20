package database

import (
	"database/sql"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"

	"api/models"
)

func GetSkin(dm *db.DatabaseManager, id uint) (*models.Skin, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}
	res := &models.Skin{}
	err = dbo.Get(res, `
		SELECT * FROM skin
		WHERE skin_id = $1`,
		id)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, ErrNotFound
		}
		return res, err
	}

	return res, nil
}

func GetAllSkins(dm *db.DatabaseManager) (*[]models.Skin, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}

	skins := &[]models.Skin{}
	err = dbo.Select(skins, `
		SELECT * FROM skin
		ORDER BY skin_id`)
	if err != nil {
		return skins, err
	}

	return skins, nil
}

func GetUserStore(dm *db.DatabaseManager, uID uint) (*models.Store, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}
	res := &models.Store{}
	err = dbo.Get(res, `
		SELECT coins, skin FROM user_profile
		WHERE user_id = $1`,
		uID)
	if err != nil {
		if err == sql.ErrNoRows {
			return res, UserNotFoundError{"id"}
		}
		return res, err
	}

	purchased, err := GetBoughtSkins(dm, uID)
	if err != nil {
		return res, err
	}
	res.PurchasedSkins = *purchased

	return res, nil
}

func GetBoughtSkins(dm *db.DatabaseManager, uID uint) (*[]uint, error) {
	dbo, err := dm.DB()
	if err != nil {
		return nil, err
	}
	res := new([]uint)
	err = dbo.Select(res, `
		SELECT skin_id FROM user_purchased_skins
		WHERE user_id = $1`,
		uID)
	if err != nil {
		return res, err
	}

	return res, nil
}

func ChangeUserCoinAmount(dm *db.DatabaseManager, uID uint, sum int) error {
	dbo, err := dm.DB()
	if err != nil {
		return err
	}
	_, err = dbo.Exec(`
		UPDATE user_profile
		SET coins = coins + $1
		WHERE user_id = $2`,
		sum, uID,
	)
	if err != nil {
		return err
	}

	return nil
}

func TxChangeUserCoinAmount(tx *sql.Tx, uID uint, sum int) error {
	_, err := tx.Exec(`
		UPDATE user_profile
		SET coins = coins + $1
		WHERE user_id = $2`,
		sum, uID,
	)
	if err != nil {
		return err
	}

	return nil
}

func BuySkin(dm *db.DatabaseManager, uID uint, skin *models.Skin) error {
	dbo, err := dm.DB()
	if err != nil {
		return err
	}
	tx, err := dbo.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`
		INSERT INTO user_purchased_skins (user_id, skin_id)
		VALUES ($1, $2)`,
		uID, skin.ID,
	)
	if err != nil {
		return err
	}

	if skin.Cost != 0 {
		err = TxChangeUserCoinAmount(tx, uID, -skin.Cost)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func ChangeSkin(dm *db.DatabaseManager, uID, skin uint) error {
	dbo, err := dm.DB()
	if err != nil {
		return err
	}
	if skin != 0 {
		_, err = dbo.Exec(`
			UPDATE user_profile
			SET skin = $1
			WHERE user_id = $2`,
			skin, uID)
	} else { // equip default skin
		_, err = dbo.Exec(`
			UPDATE user_profile
			SET skin = NULL
			WHERE user_id = $1`,
			uID)
	}
	if err != nil {
		return err
	}

	return nil
}
