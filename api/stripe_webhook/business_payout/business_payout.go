package sw_payout

import (
	"database/sql"
	"errors"

	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/errors/bus_errors"
	"github.com/johnyeocx/usual/server/errors/gen_errors"
	"github.com/stripe/stripe-go/v74"
)

func InsertPayout(sqlDB *sql.DB, sp stripe.Payout) (*models.RequestError) {
	
	// Get business id 
	b := busdb.BusinessDB{DB: sqlDB}
	p := busdb.PayoutDB{DB: sqlDB}

	var bus models.Business
	var extAcctId int
	if sp.Type == stripe.PayoutTypeBank {
		business, baId, err := b.GetBusinessFromBankStripeID(sp.Destination.ID)
		if err != nil {
			return bus_errors.GetBusFailedReqErr(err)
		}

		bus = *business
		extAcctId = *baId
	} else {
		return gen_errors.NotSupportedReqErr(errors.New("card payout not currently supported"))
	}

	// 2. Insert into db
	err := p.InsertPayout(bus.ID, extAcctId, sp)
	if err != nil {
		return bus_errors.InsertPayoutFailedErr(err)
	}

	return nil
}

