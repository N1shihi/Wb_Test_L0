-- 000001_create_orders_tables.up.sql

CREATE TABLE IF NOT EXISTS orders (
                                      order_uid VARCHAR PRIMARY KEY,
                                      track_number VARCHAR,
                                      entry VARCHAR,
                                      locale VARCHAR,
                                      internal_signature VARCHAR,
                                      customer_id VARCHAR,
                                      delivery_service VARCHAR,
                                      shardkey VARCHAR,
                                      sm_id INTEGER,
                                      date_created TIMESTAMP,
                                      oof_shard VARCHAR
);

CREATE TABLE IF NOT EXISTS deliveries (
                                          order_uid VARCHAR PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
                                          name VARCHAR,
                                          phone VARCHAR,
                                          zip VARCHAR,
                                          city VARCHAR,
                                          address VARCHAR,
                                          region VARCHAR,
                                          email VARCHAR
);

CREATE TABLE IF NOT EXISTS payments (
                                        transaction VARCHAR PRIMARY KEY,
                                        order_uid VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
                                        request_id VARCHAR,
                                        currency VARCHAR,
                                        provider VARCHAR,
                                        amount INTEGER,
                                        payment_dt BIGINT,
                                        bank VARCHAR,
                                        delivery_cost INTEGER,
                                        goods_total INTEGER,
                                        custom_fee INTEGER
);

CREATE TABLE IF NOT EXISTS items (
                                     id SERIAL PRIMARY KEY,
                                     order_uid VARCHAR REFERENCES orders(order_uid) ON DELETE CASCADE,
                                     chrt_id INTEGER,
                                     track_number VARCHAR,
                                     price INTEGER,
                                     rid VARCHAR,
                                     name VARCHAR,
                                     sale INTEGER,
                                     size VARCHAR,
                                     total_price INTEGER,
                                     nm_id INTEGER,
                                     brand VARCHAR,
                                     status INTEGER
);
