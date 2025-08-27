package repository

import (
	"Wb_Test_L0/internal/models"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveOrder(o models.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var t time.Time
	if o.DateCreated != "" {
		if parsed, perr := time.Parse(time.RFC3339, o.DateCreated); perr == nil {
			t = parsed
		}
	}

	ordersQuery := `
INSERT INTO orders(order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (order_uid) DO UPDATE SET
track_number=EXCLUDED.track_number, entry=EXCLUDED.entry, locale=EXCLUDED.locale,
internal_signature=EXCLUDED.internal_signature, customer_id=EXCLUDED.customer_id,
delivery_service=EXCLUDED.delivery_service, shardkey=EXCLUDED.shardkey, sm_id=EXCLUDED.sm_id,
date_created=EXCLUDED.date_created, oof_shard=EXCLUDED.oof_shard
`
	if _, err = tx.Exec(ordersQuery, o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID, o.DeliveryService, o.Shardkey, o.SmID, t, o.OofShard); err != nil {
		return err
	}

	delQ := `
INSERT INTO deliveries(order_uid, name, phone, zip, city, address, region, email)
VALUES($1,$2,$3,$4,$5,$6,$7,$8)
ON CONFLICT (order_uid) DO UPDATE SET
name=EXCLUDED.name, phone=EXCLUDED.phone, zip=EXCLUDED.zip, city=EXCLUDED.city,
address=EXCLUDED.address, region=EXCLUDED.region, email=EXCLUDED.email
`
	if _, err = tx.Exec(delQ, o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email); err != nil {
		return err
	}

	payQ := `
INSERT INTO payments(transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (transaction) DO UPDATE SET
order_uid=EXCLUDED.order_uid, request_id=EXCLUDED.request_id, currency=EXCLUDED.currency,
provider=EXCLUDED.provider, amount=EXCLUDED.amount, payment_dt=EXCLUDED.payment_dt,
bank=EXCLUDED.bank, delivery_cost=EXCLUDED.delivery_cost, goods_total=EXCLUDED.goods_total, custom_fee=EXCLUDED.custom_fee
`
	if _, err = tx.Exec(payQ, o.Payment.Transaction, o.OrderUID, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee); err != nil {
		return err
	}

	if _, err = tx.Exec("DELETE FROM items WHERE order_uid = $1", o.OrderUID); err != nil {
		return err
	}
	itQ := `
INSERT INTO items(order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
`
	for _, it := range o.Items {
		if _, err = tx.Exec(itQ, o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.Rid, it.Name, it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetOrderByID(id string) (*models.Order, error) {
	q := `SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created::text, oof_shard FROM orders WHERE order_uid = $1`
	row := r.db.QueryRow(q, id)
	var o models.Order
	var dateStr sql.NullString
	if err := row.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID, &o.DeliveryService, &o.Shardkey, &o.SmID, &dateStr, &o.OofShard); err != nil {
		return nil, err
	}
	if dateStr.Valid {
		o.DateCreated = dateStr.String
	}

	dq := `SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE order_uid = $1`
	dr := r.db.QueryRow(dq, id)
	var d models.Delivery
	if err := dr.Scan(&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email); err == nil {
		o.Delivery = d
	}

	pq := `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee FROM payments WHERE order_uid = $1 LIMIT 1`
	pr := r.db.QueryRow(pq, id)
	var p models.Payment
	if err := pr.Scan(&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDt, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee); err == nil {
		o.Payment = p
	}

	itemsQ := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1`
	rows, err := r.db.Query(itemsQ, id)
	if err == nil {
		defer rows.Close()
		var items []models.Item
		for rows.Next() {
			var it models.Item
			if err := rows.Scan(&it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid, &it.Name, &it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status); err == nil {
				items = append(items, it)
			}
		}
		o.Items = items
	}

	return &o, nil
}

func (r *Repository) GetLastOrders(limit int) ([]models.Order, error) {
	q := `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created::text, oof_shard
		FROM orders 
		ORDER BY date_created DESC 
		LIMIT $1`
	rows, err := r.db.Query(q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var dateStr sql.NullString
		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
			&o.CustomerID, &o.DeliveryService, &o.Shardkey, &o.SmID, &dateStr, &o.OofShard); err != nil {
			return nil, err
		}
		if dateStr.Valid {
			o.DateCreated = dateStr.String
		}

		dq := `SELECT name, phone, zip, city, address, region, email FROM deliveries WHERE order_uid = $1`
		dr := r.db.QueryRow(dq, o.OrderUID)
		var d models.Delivery
		if err := dr.Scan(&d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email); err == nil {
			o.Delivery = d
		}

		pq := `SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee 
		       FROM payments WHERE order_uid = $1 LIMIT 1`
		pr := r.db.QueryRow(pq, o.OrderUID)
		var p models.Payment
		if err := pr.Scan(&p.Transaction, &p.RequestID, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDt,
			&p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee); err == nil {
			o.Payment = p
		}

		itemsQ := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status 
		           FROM items WHERE order_uid = $1`
		itemRows, err := r.db.Query(itemsQ, o.OrderUID)
		if err == nil {
			defer itemRows.Close()
			var items []models.Item
			for itemRows.Next() {
				var it models.Item
				if err := itemRows.Scan(&it.ChrtID, &it.TrackNumber, &it.Price, &it.Rid, &it.Name,
					&it.Sale, &it.Size, &it.TotalPrice, &it.NmID, &it.Brand, &it.Status); err == nil {
					items = append(items, it)
				}
			}
			o.Items = items
		}

		orders = append(orders, o)
	}

	return orders, nil
}
