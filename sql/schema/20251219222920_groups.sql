-- +goose Up
-- +goose StatementBegin
CREATE TABLE groups (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE group_packages (
    group_id   INTEGER NOT NULL,
    package_id INTEGER NOT NULL,

    PRIMARY KEY (group_id, package_id),

    FOREIGN KEY (group_id)
        REFERENCES groups(id)
        ON DELETE CASCADE,

    FOREIGN KEY (package_id)
        REFERENCES packages(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_groups_name ON groups(name);
CREATE INDEX idx_group_packages_group ON group_packages(group_id);
CREATE INDEX idx_group_packages_package ON group_packages(package_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_packages;
DROP TABLE IF EXISTS groups;
-- +goose StatementEnd
