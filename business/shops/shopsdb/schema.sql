CREATE TABLE shops (
  id   text PRIMARY KEY,
  user_id text NOT NULL,
  name text    NOT NULL
);

CREATE INDEX shops_by_user ON shops (user_id);

