-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS credit_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    amount DECIMAL(15,2) NOT NULL,
    term INT NOT NULL,
    status VARCHAR(20) NOT NULL,
    reject_reason TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS credit_applications;
-- +goose StatementEnd
