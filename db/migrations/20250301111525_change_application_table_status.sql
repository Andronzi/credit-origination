-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN status_temp INT;

UPDATE credit_applications 
SET status_temp = CASE status
    WHEN 'NEW' THEN 0
    WHEN 'SCORING' THEN 1
    WHEN 'EMPLOYMENT_CHECK' THEN 2
    WHEN 'APPROVED' THEN 3
    WHEN 'REJECTED' THEN 4
    ELSE 0
END;

ALTER TABLE credit_applications 
DROP COLUMN status;

ALTER TABLE credit_applications 
RENAME COLUMN status_temp TO status;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN status_temp VARCHAR(20);

UPDATE credit_applications 
SET status_temp = CASE status
    WHEN 0 THEN 'NEW'
    WHEN 1 THEN 'SCORING'
    WHEN 2 THEN 'EMPLOYMENT_CHECK'
    WHEN 3 THEN 'APPROVED'
    WHEN 4 THEN 'REJECTED'
    ELSE 'NEW'
END;

ALTER TABLE credit_applications 
DROP COLUMN status;

ALTER TABLE credit_applications 
RENAME COLUMN status_temp TO status;
-- +goose StatementEnd
