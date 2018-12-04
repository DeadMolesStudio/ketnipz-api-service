package database

import (
	"api/models"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"
)

func GetUserPositionsDescendingPaginated(dm *db.DatabaseManager, p *models.FetchScoreboardPage) (
	*[]models.Position, int, error) {
	total, err := GetCountOfUsers(dm)
	if err != nil {
		return nil, total, err
	}

	// TODO: optimize it
	dbo, err := dm.DB()
	if err != nil {
		return nil, total, err
	}

	records := &[]models.Position{}
	err = dbo.Select(records, `
		SELECT user_id, nickname, record FROM user_profile
		ORDER BY record DESC
		LIMIT $1
		OFFSET $2`,
		p.Limit, p.Limit*p.Page)
	if err != nil {
		return records, total, err
	}

	return records, total, nil
}
