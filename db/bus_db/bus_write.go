package busdb

import "database/sql"
type BusinessDB struct {
	DB	*sql.DB
}

func (businessDB *BusinessDB) CreateBusinessProfile(
	id int,
	businessCategory string, 
	businessUrl string, 
	individualId int,
) (error) {
	updateStatement := `UPDATE business SET business_category=$1, business_url=$2, individual_id=$3
	WHERE business_id=$4 AND email_verified=true`

	_, err := businessDB.DB.Exec(updateStatement, 
		businessCategory, 
		businessUrl, 
		individualId, 
		id,
	)
	if err != nil {
		return err
	}

	return nil
}