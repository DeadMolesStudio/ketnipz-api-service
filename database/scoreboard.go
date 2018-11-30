package database

import (
	"api/models"

	db "github.com/go-park-mail-ru/2018_2_DeadMolesStudio/database"
)

func GetUserPositionsDescendingPaginated(dm *db.DatabaseManager, p *models.FetchScoreboardPage) (
	[]models.Position, int, error) {
	records := []models.Position{}
	total, err := GetCountOfUsers(dm)
	if err != nil {
		return records, total, err
	}

	// TODO: optimize it
	dbo, err := dm.DB()
	if err != nil {
		return records, total, err
	}

	rows, err := dbo.Queryx(`
		SELECT user_id, nickname, record FROM user_profile
		ORDER BY record DESC
		LIMIT $1
		OFFSET $2`,
		p.Limit, p.Limit*p.Page)
	if err != nil {
		return records, total, err
	}

	r := models.Position{}
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return records, total, err
		}
		err := rows.StructScan(&r)
		if err != nil {
			return records, total, err
		}
		records = append(records, r)
	}

	return records, total, nil
}
