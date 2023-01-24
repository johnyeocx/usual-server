package c_business

import (
	"database/sql"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
)

func GetExploreData(sqlDB *sql.DB) ([]models.ExploreResult, error) {
	b := db.BusinessDB{DB: sqlDB}

	res, err := b.GetBusinessWithTopSubbedProduct()
	if err != nil {
		return nil, err
	}

	return res, nil
}


func SearchAccounts(sqlDB *sql.DB, query string) ([]models.Business, error) {
	b := db.BusinessDB{DB: sqlDB}
	accounts, err := b.SearchAccounts(query)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

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


func GetBusinessSubProducts(sqlDB *sql.DB, businessId int) (map[string]interface{}, error) {

	b := db.BusinessDB{DB: sqlDB}

	businessDB := db.BusinessDB{DB: sqlDB}
	business, err := businessDB.GetBusinessByIDWithSubCount(businessId)
	if err != nil {
		return nil, err
	}

	// 2. Get products
	subProducts, err := b.GetCBusinessSubProducts(businessId)
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