-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN product_code UUID,
ADD COLUMN product_version VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications 
DROP COLUMN product_code,
DROP COLUMN product_version;
-- +goose StatementEnd
