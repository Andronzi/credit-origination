-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications
ALTER COLUMN to_bank_account_id TYPE UUID USING to_bank_account_id::UUID;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications
ALTER COLUMN to_bank_account_id TYPE VARCHAR USING to_bank_account_id::VARCHAR;
-- +goose StatementEnd
