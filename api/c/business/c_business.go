package c_business

import (
	"database/sql"
	"net/http"

	"github.com/johnyeocx/usual/server/db"
	busdb "github.com/johnyeocx/usual/server/db/bus_db"
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

func SearchSubProducts(sqlDB *sql.DB, query string) ([]models.ExploreResult, error) {
	b := db.BusinessDB{DB: sqlDB}
	res, err := b.SearchSubProducts(query)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func GetBusiness(sqlDB *sql.DB, businessId int) (*models.Business, *models.RequestError) {
	b := busdb.BusinessDB{DB: sqlDB}

	// 1. Get business
	business, err := b.GetBusinessByID(businessId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return business, nil
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

func GetCSubProduct(sqlDB *sql.DB, productId int) (*models.SubscriptionProduct, error) {

	b := db.BusinessDB{DB: sqlDB}

	subProduct, err := b.GetCSubProduct(productId)
	
	if err != nil {
		return nil, err
	}

	// 2. Get products
	usages, err := b.GetSubProductUsages(subProduct.Product.ProductID)
	if err != nil {
		return nil, err
	}

	subProduct.SubPlan.Usages = usages


	return subProduct, nil
}