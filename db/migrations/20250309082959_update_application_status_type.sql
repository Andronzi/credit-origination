-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN status_temp VARCHAR(50);

UPDATE credit_applications 
SET status_temp = CASE 
    WHEN status = 'DRAFT' THEN 'DRAFT'
    WHEN status = 'APPLICATION' THEN 'APPLICATION_CREATED'
    WHEN status = 'SCORING' THEN 'SCORING'
    WHEN status = 'EMPLOYMENT_CHECK' THEN 'EMPLOYMENT_CHECK'
    WHEN status = 'APPROVED' THEN 'APPROVED'
    WHEN status = 'REJECTED' THEN 'REJECTED'
    ELSE 'DRAFT'
END;

ALTER TABLE credit_applications 
DROP COLUMN status;

ALTER TABLE credit_applications 
RENAME COLUMN status_temp TO status;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN status_temp VARCHAR(50);

UPDATE credit_applications 
SET status_temp = CASE 
    WHEN status = 'DRAFT' THEN 'DRAFT'
    WHEN status = 'APPLICATION_CREATED' THEN 'APPLICATION'
    WHEN status = 'APPLICATION_AGREEMENT_CREATED' THEN 'APPLICATION'
    WHEN status = 'SCORING' THEN 'SCORING'
    WHEN status = 'EMPLOYMENT_CHECK' THEN 'EMPLOYMENT_CHECK'
    WHEN status = 'APPROVED' THEN 'APPROVED'
    WHEN status = 'REJECTED' THEN 'REJECTED'
    ELSE 'DRAFT'
END;

ALTER TABLE credit_applications 
DROP COLUMN status;

ALTER TABLE credit_applications 
RENAME COLUMN status_temp TO status;
-- +goose StatementEnd
