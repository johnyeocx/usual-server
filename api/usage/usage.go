package usage

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
)

// func GetBusinessUsages(
// 	sqlDB *sql.DB,
// 	businessId int,
// ) (map[string]interface{}, *models.RequestError) {

// }

func InsertCusUsage(
	sqlDB *sql.DB,
	cusUuid string,
	businessId int,
	subUsageId int,
) (map[string]interface{}, *models.RequestError)  {

	u := db.UsageDB{DB: sqlDB}

	// check that business owns sub usage id
	subUsage, usageCount,  err := u.InsertCusUsageValid(cusUuid, subUsageId, businessId)
	if err == sql.ErrNoRows {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusForbidden,
		}
	} else if err != nil{
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	if !subUsage.Unlimited && int(subUsage.Amount.Int16) <= *usageCount {
		// insert into db new usage
		return map[string]interface{} {
			"overflow": true,
			"cus_usage": nil,
		}, nil
	}

	newUsage, err := u.InsertNewCusUsage(models.CusUsage{
		CusUUID: cusUuid,
		Created: time.Now(),
		SubUsageID: subUsageId,
	})

	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	// return true
	return map[string]interface{} {
		"overflow": false,
		"cus_usage": newUsage,
	}, nil
}

func ScanCusQR(
	sqlDB *sql.DB,
	cusUuid string,
	businessId int,
) (map[string]interface{}, *models.RequestError) {

	u := db.UsageDB{DB: sqlDB}
	usageInfos, err := u.GetCusUsagesOnBusiness(cusUuid, businessId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	if len(usageInfos) == 1 {
		info := usageInfos[0]
		if info.SubUsage.Unlimited || int(info.SubUsage.Amount.Int16) > info.UsageCount {
			// insert into db new usage
			newUsage, err := u.InsertNewCusUsage(models.CusUsage{
				CusUUID: cusUuid,
				Created: time.Now(),
				SubUsageID: info.SubUsage.ID,
			})

			if err != nil {
				return nil, &models.RequestError{
					Err: err,
					StatusCode: http.StatusBadGateway,
				}
			}
			
			// return true
			return map[string]interface{} {
				"only_one": true,
				"overflow": false,
				"cus_usage": newUsage,
				"usage_infos": usageInfos,
			}, nil
		} else {
			return map[string]interface{} {
				"only_one": true,
				"overflow": true,
				"cus_usage": nil,
				"usage_infos": usageInfos,
			}, nil
		}
	}

	return map[string]interface{} {
		"only_one": false,
		"overflow": nil,
		"cus_usage": nil,
		"usage_infos": usageInfos,
	}, nil
}