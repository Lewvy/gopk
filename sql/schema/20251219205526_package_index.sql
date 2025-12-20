-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_last_used ON packages(last_used);
CREATE INDEX idx_freq ON packages(freq);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_last_used;
-- +goose StatementEnd
