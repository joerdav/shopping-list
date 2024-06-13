CREATE TABLE items (
	id   text primary key,
	name text    not null,
	user_id text not null,
	shop_id text not null
);

CREATE TABLE lists (
	id   text primary key,
	user_id text not null,
	created_date integer not null,
	FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX lists_by_user ON lists (user_id);

CREATE TABLE list_items (
	item_id text not null,
	list_id text not null,
	quantity integer not null,
	FOREIGN KEY (list_id) REFERENCES lists(id),
	PRIMARY KEY (item_id, list_id)
);

CREATE TABLE list_recipes (
	recipe_id text not null,
	list_id text not null,
	quantity integer not null,
	FOREIGN KEY (list_id) REFERENCES lists(id),
	PRIMARY KEY (recipe_id, list_id)
);

CREATE TABLE recipes (
  id   text PRIMARY KEY,
  user_id text not null,
  name text    NOT NULL
);

CREATE TABLE ingredients (
	item_id text NOT NULL,
	recipe_id text NOT NULL,
	quantity integer NOT NULL,
	PRIMARY KEY (item_id, recipe_id)
);

CREATE TABLE shops (
  id   text PRIMARY KEY,
  user_id text NOT NULL,
  name text    NOT NULL
);

CREATE INDEX shops_by_user ON shops (user_id);

CREATE TABLE users (
  id   text PRIMARY KEY
);
