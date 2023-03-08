package db

import (
	"database/sql"
	"time"

	"github.com/johnyeocx/usual/server/db/models"
)

type UsageDB struct {
	DB *sql.DB
}

func (u *UsageDB) InsertCusUsageValid(
	cusUuid string,
	subUsageId int,
	businessId int,
) (*models.SubUsage, *int, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	weekday := now.Weekday()
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day() - int(weekday - 1), 0, 0, 0, 0, time.UTC)

	startOfMonth := time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.UTC)
	startOfYear := time.Date(now.Year(), 0, 0, 0, 0, 0, 0, time.UTC)

	stmt := `
		SELECT su.unlimited, su.interval, su.amount, COUNT(cu.sub_usage_id) as usage_count 
		from customer as c 
		JOIN subscription as s on c.customer_id=s.customer_id
		JOIN subscription_plan as sp ON s.plan_id=sp.plan_id
		JOIN subscription_usage as su ON su.plan_id=sp.plan_id
		JOIN product as p on p.product_id=sp.product_id
		JOIN business as b ON b.business_id=p.business_id

		LEFT JOIN customer_usage as cu ON cu.sub_usage_id=su.sub_usage_id AND (
			su."interval" = 'day' AND cu.created > $1 OR
			su."interval" = 'week' AND cu.created > $2 OR
			su."interval" = 'month' AND cu.created > $3 OR
			su."interval" = 'year' AND cu.created > $4
		)

		WHERE c.uuid=$5 AND b.business_id=$6 AND su.sub_usage_id=$7
		GROUP BY 
		cu.sub_usage_id, c.customer_id, s.plan_id, p.product_id, su.sub_usage_id
	`

	subUsage := models.SubUsage{}
	var usageCount int
	err := u.DB.QueryRow(
		stmt, startOfDay, startOfWeek, startOfMonth, startOfYear, 
		cusUuid, businessId, subUsageId).Scan(
		&subUsage.Unlimited,
		&subUsage.Interval, 
		&subUsage.Amount,
		&usageCount,
	)

	if err != nil {
		return nil, nil, err
	}
	return &subUsage, &usageCount, nil
}

func (u *UsageDB) GetCusUsagesOnBusiness(
	cusUuid string, 
	busId int,
) ([]models.UsageInfo, error){

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	weekday := now.Weekday()
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day() - int(weekday - 1), 0, 0, 0, 0, time.UTC)

	startOfMonth := time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, time.UTC)
	startOfYear := time.Date(now.Year(), 0, 0, 0, 0, 0, 0, time.UTC)



	query := `
		SELECT 
		c.uuid, c.first_name, c.last_name,
		s.plan_id, 
		su.title, su.sub_usage_id, su.unlimited, su.interval, su.amount,
		p.product_id, p.name, 
		COUNT(cu.sub_usage_id) as usage_count 
		from 
		customer as c z
		JOIN subscription as s on c.customer_id=s.customer_id
		JOIN subscription_plan as sp ON s.plan_id=sp.plan_id
		JOIN subscription_usage as su ON su.plan_id=sp.plan_id
		JOIN product as p on p.product_id=sp.product_id
		JOIN business as b ON b.business_id=p.business_id

		LEFT JOIN customer_usage as cu ON cu.sub_usage_id=su.sub_usage_id AND (
			su."interval" = 'day' AND cu.created > $1 OR
			su."interval" = 'week' AND cu.created > $2 OR
			su."interval" = 'month' AND cu.created > $3 OR
			su."interval" = 'year' AND cu.created > $4
		)

		WHERE c.uuid=$5 AND b.business_id=$6
		GROUP BY 
		cu.sub_usage_id, c.customer_id, s.plan_id, p.product_id, su.sub_usage_id
		ORDER BY su.sub_usage_id
	`

	rows, err := u.DB.Query(
		query, startOfDay, startOfWeek, startOfMonth, startOfYear, 
		cusUuid, busId,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	usageInfos := []models.UsageInfo{}
	for rows.Next() {
		var info models.UsageInfo
		if err := rows.Scan(
			&info.CusUUID,
			&info.CusFirstName,
			&info.CusLastName,
			&info.PlanID,
			&info.SubUsage.Title,
			&info.SubUsage.ID,
			&info.SubUsage.Unlimited,
			&info.SubUsage.Interval,
			&info.SubUsage.Amount,
			&info.ProductID,
			&info.ProductName,
			&info.UsageCount,
		); err != nil {
			continue
		}

		usageInfos = append(usageInfos, info)
	}

	return usageInfos, nil
}

func (u *UsageDB) InsertNewCusUsage(
	usage models.CusUsage,
) (*models.CusUsage, error){

	query := `INSERT into customer_usage (customer_uuid, created, sub_usage_id) 
		VALUES ($1, $2, $3) RETURNING usage_id, customer_uuid, created, sub_usage_id
	`

	returnedUsage := models.CusUsage{}
	err := u.DB.QueryRow(
		query, 
		usage.CusUUID, 
		usage.Created, 
		usage.SubUsageID,
	).Scan(
		&returnedUsage.ID, 
		&returnedUsage.CusUUID,
		&returnedUsage.Created,
		&returnedUsage.SubUsageID,
	)

	if err != nil {
		return nil, err
	}

	return &usage, nil
}