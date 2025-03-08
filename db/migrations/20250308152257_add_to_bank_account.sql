-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN to_bank_account_id UUID;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications DROP COLUMN to_bank_account_id;
-- +goose StatementEnd
