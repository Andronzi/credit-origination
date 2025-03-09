-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ALTER COLUMN interest TYPE DECIMAL(15,2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications 
ALTER COLUMN interest TYPE FLOAT USING interest::FLOAT;
-- +goose StatementEnd
