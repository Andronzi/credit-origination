-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN interest FLOAT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications 
DROP COLUMN interest;
-- +goose StatementEnd
