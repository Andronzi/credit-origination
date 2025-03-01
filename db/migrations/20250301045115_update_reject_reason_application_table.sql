-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ALTER COLUMN reject_reason TYPE TEXT,
ALTER COLUMN reject_reason DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE credit_applications 
SET reject_reason = ''
WHERE reject_reason IS NULL;

ALTER TABLE credit_applications 
ALTER COLUMN reject_reason SET NOT NULL,
ALTER COLUMN reject_reason TYPE VARCHAR(255);
-- +goose StatementEnd
