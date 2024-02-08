CREATE TABLE recipes (
  id   text PRIMARY KEY,
  name text    NOT NULL
);

CREATE TABLE ingredients (
	item_id text NOT NULL,
	recipe_id text NOT NULL,
	quantity integer NOT NULL,
	PRIMARY KEY (item_id, recipe_id)
);
