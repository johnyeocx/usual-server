package c_business

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db"
)

func GetBusinessByID(sqlDB *sql.DB, businessId int) (map[string]interface{}, error) {

	b := db.BusinessDB{DB: sqlDB}

	// 1. Get business
	business, err := b.GetBusinessByID(businessId)
	if err != nil {
		return nil, err
	}

	// 2. Get products
	subProducts, err := b.GetBusinessSubProducts(businessId)
	if err != nil {
		return nil, err
	}

	// 3. Get product categories
	productCats, err := b.GetBusinessProductCategories(businessId)
	if err != nil {
		return nil, err
	}

	return map[string] interface{} {
		"business": business,
		"sub_products": *subProducts,
		"product_categories": *productCats,
	}, nil
}