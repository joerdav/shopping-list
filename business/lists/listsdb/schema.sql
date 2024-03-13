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
