CREATE TABLE IF NOT EXISTS courier (
    id serial primary key,
    type varchar(10),
    created_at timestamp without time zone DEFAULT now()
);


CREATE TABLE IF NOT EXISTS courier_regions (
    id serial primary key,
    courier_id bigint,
    number integer,
    foreign key (courier_id) REFERENCES courier (id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS courier_working_hours (
    id serial primary key,
    courier_id bigint REFERENCES courier (id) ON DELETE CASCADE NOT NULL,
    starts time without time zone,
    ends time without time zone
);

CREATE TABLE IF NOT EXISTS group_order (
    id serial primary key,
    courier_id bigint REFERENCES courier (id) ON DELETE CASCADE NOT NULL,
    date timestamp without time zone
);

CREATE TABLE IF NOT EXISTS "order" (
    id serial primary key,
    cost integer,
    weight numeric,
    region integer,
    group_id bigint REFERENCES group_order (id) ON DELETE SET NULL,
    completed_time timestamp without time zone,
    created_at timestamp without time zone DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_courier (
    order_id bigint NOT NULL UNIQUE,
    courier_id bigint NOT NULL,
    completed_time timestamp without time zone,
    PRIMARY KEY (order_id, courier_id)
);

CREATE TABLE IF NOT EXISTS order_delivery_hours (
    id serial primary key,
    order_id bigint REFERENCES "order" (id) ON DELETE CASCADE NOT NULL,
    starts time without time zone,
    ends time without time zone
);


CREATE INDEX IF NOT EXISTS idx_courier_type ON courier USING btree (type);

CREATE INDEX IF NOT EXISTS idx_order_completed_time ON "order" USING btree (completed_time);

CREATE INDEX IF NOT EXISTS idx_order_courier_completed_time ON order_courier USING btree (completed_time);
