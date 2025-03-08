-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ALTER COLUMN product_code TYPE VARCHAR USING product_code::VARCHAR;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications
ALTER COLUMN product_code TYPE UUID USING product_code::UUID;
-- +goose StatementEnd
