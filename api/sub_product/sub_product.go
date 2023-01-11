package sub_product

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db"
)


func GetSubProductStats(
	sqlDB *sql.DB, 
	businessId int, 
	productId int,
) (error) {
	b := db.BusinessDB{DB: sqlDB}

	// 1. check that business owns product
	err := b.BusinessOwnsProduct(businessId, productId)
	if err != nil {
		return err
	}

	// 2. retrieve list of subscriptions for product
	b.GetSubProductSubscriptions(productId)

	// 3. for list of subscriptions, retrieve stripe invoice data

	return nil
}