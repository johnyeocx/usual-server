package sub_product

import (
	"database/sql"
	"net/http"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
)


func GetSubProductStats(
	sqlDB *sql.DB, 
	businessId int, 
	productId int,
) (map[string]interface{}, *models.RequestError) {
	b := db.BusinessDB{DB: sqlDB}

	// 1. check that business owns product
	err := b.BusinessOwnsProduct(businessId, productId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	// 2. retrieve list of subscribers for product
	subscribers, err := b.GetSubProductSubscribers(productId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. retrieve list of invoice data for product
	invoices, err := b.GetSubProductInvoices(productId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}


	return map[string]interface{}{
		"subscribers": subscribers,
		"invoices": invoices,
	}, nil
}