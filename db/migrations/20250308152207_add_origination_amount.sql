-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN origination_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications DROP COLUMN origination_amount;
-- +goose StatementEnd
