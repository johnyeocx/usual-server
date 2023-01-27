package sub_product

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)


func createSubProduct (
	sqlDB *sql.DB,
	businessId int,
	category *models.ProductCategory,
	product *models.Product, 
	subPlan *models.SubscriptionPlan,
	usages []models.SubUsage,
) (*int, *models.SubscriptionProduct, error) {

	
	db := db.BusinessDB{DB: sqlDB}

	// 1. Insert new category id
	var newCatId *int;
	var catId = category.CategoryID
	
	if (category.CategoryID == nil) {
		id, err := db.InsertProductCategory(businessId, category.Title)
		if err != nil {
			return nil, nil, err
		}


		newCatId = id;
		catId = id;
	}

	stripeProductId, stripePriceId, err := my_stripe.CreateNewSubProduct(product.Name, *subPlan)
	if err != nil {
		return nil, nil, err
	}

	
	// 1. insert subscription
	insertedProduct, err := db.InsertProduct(businessId, catId, product, *stripeProductId)
	if err != nil {
		return nil, nil, err
	}

	// 2. insert subscription
	insertedPlan, err := db.InsertSubPlan(insertedProduct.ProductID, subPlan, *stripePriceId)
	if err != nil {
		return nil, nil, err
	}

	// 3. insert usages
	insertedUsages, err := db.InsertUsages(insertedPlan.PlanID, usages)
	if err != nil {
		return nil, nil, err
	}

	insertedPlan.Usages = &insertedUsages

	if err != nil {
		return nil, nil, err
	}
	

	return newCatId, &models.SubscriptionProduct{
		Product: *insertedProduct,
		SubPlan: *insertedPlan,
	}, nil
}

func GetBusinessProducts(
	sqlDB *sql.DB,
	businessId int,
) (*[]models.ProductCategory, *[]models.SubscriptionProduct, error) {

	b := db.BusinessDB{DB: sqlDB}
	productCategories, err := b.GetBusinessProductCategories(businessId)
	if err != nil {
		return nil, nil,err
	}

	subProducts, err := b.GetBusinessSubProducts(businessId)
	if err != nil {
		return nil,nil, err
	}

	return productCategories, subProducts, nil
}


func GetSubProductStats(
	sqlDB *sql.DB, 
	businessId int, 
	productId int,
) (map[string]interface{}, *models.RequestError) {
	b := db.BusinessDB{DB: sqlDB}

	// 1. check that business owns product
	_, err := b.BusinessOwnsProduct(businessId, productId)
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


func UpdateProductName(
	sqlDB *sql.DB,
	businessId int,
	productId int,
	newName string,
) (*models.RequestError) {
	b := db.BusinessDB{DB: sqlDB}

	// 1. Business owns product
	product, err := b.BusinessOwnsProduct(businessId, productId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	err = my_stripe.UpdateProductName(*product.StripeProductID, newName)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = b.SetProductName(businessId, productId, newName)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}

func UpdateProductCategory(
	sqlDB *sql.DB,
	businessId int,
	productId int,
	title string,
	catId *int,
) (*int, *models.RequestError) {
	b := db.BusinessDB{DB: sqlDB}

	// Get previous product category
	prevCatId, err := b.GetProductCatId(productId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	catId, err = b.SetProductCategory(businessId, productId, catId, title)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	b.DeleteCategoryIfEmpty(*prevCatId)

	if catId != nil {
		return catId, nil
	} else {
		return nil, nil
	}
}

func DeleteSubProduct(
	sqlDB *sql.DB,
	s3Sess *session.Session,
	businessId int,
	productId int,
) (*models.RequestError) {
	b := db.BusinessDB{DB: sqlDB}

	// 1. Check that business owns product
	_, err := b.BusinessOwnsProduct(businessId, productId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusUnauthorized,
		}
	}

	// 2. Get delete data
	data, err := b.GetSubProductDeleteData(productId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}


	// 3. Delete all stripe_subs
	stripeSubIds := (*data)["stripe_sub_ids"].([]string)

	
	for _, id := range stripeSubIds {

		err := my_stripe.CancelSubscription(id)

		if err != nil {
			return &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}

		err = b.DeleteSubscription(id)
		if err != nil {
			return &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}
	}

	// 4. Delete stripe product & price
	stripePriceId := (*data)["stripe_price_id"].(string)
	stripeProductId := (*data)["stripe_product_id"].(string)
	err = my_stripe.DisableProduct(stripeProductId, stripePriceId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 5. Delete product from DB
	err = b.DeleteSubProduct(productId, (*data)["plan_id"].(int))
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	key := "./business/product_image/" + strconv.Itoa(productId)
	cloud.DeleteImage(s3Sess, key)

	// 6. Delete category
	catId := (*data)["category_id"].(int)
	b.DeleteCategoryIfEmpty(catId)
	
	return nil
}