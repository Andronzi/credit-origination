-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications RENAME COLUMN amount TO disbursement_amount;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications RENAME COLUMN disbursement_amount TO amount;
-- +goose StatementEnd
