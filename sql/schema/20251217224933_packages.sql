-- +goose Up
-- +goose StatementBegin
CREATE TABLE packages (
	id integer PRIMARY KEY AUTOINCREMENT ,
	name TEXT UNIQUE NOT NULL,
	url TEXT UNIQUE NOT NULL,
	version TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE packages;
-- +goose StatementEnd
