package business

import (
	"database/sql"

	busdb "github.com/johnyeocx/usual/server/db/bus_db"
)

func getBusinessStats(
	sqlDB *sql.DB,
	businessId int,
) (map[string]interface{}, error) {
	b := busdb.BusinessDB{DB: sqlDB}
	subInfos, err := b.GetBusinessSubs(businessId)
	if err != nil {
		return nil, err
	}

	// business accounts
	bankAccounts, err := b.GetBusinessBankAccounts(businessId)
	if err != nil {
		return nil, err
	}

	invoices, err := b.GetBusinessInvoices(businessId, 20)
	if err != nil {
		return nil, err
	}

	usageInfos, err := b.GetBusinessUsages(businessId, 20)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"sub_infos": subInfos,
		"invoices": invoices,
		"usage_infos": usageInfos,
		"bank_accounts": bankAccounts,
	}, nil
}

func GetTotalAndPayouts (
	sqlDB *sql.DB, 
	busId int,
) (map[string]interface{}, error){
	// get avail payout
	// connect bank account

	// total received
	b := busdb.BusinessDB{DB: sqlDB}
	total, err :=b.GetBusinessTotalReceived(busId) 
	if err != nil {
		return nil, err
	} 

	return map[string]interface{}{
		"total": total,
	}, err
}