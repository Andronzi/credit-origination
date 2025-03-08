-- +goose Up
-- +goose StatementBegin
ALTER TABLE credit_applications 
ADD COLUMN status_temp VARCHAR(20);

UPDATE credit_applications 
SET status_temp = CASE 
    WHEN status = 0 THEN 'DRAFT'
    WHEN status = 1 THEN 'SCORING'
    WHEN status = 2 THEN 'EMPLOYMENT_CHECK'
    WHEN status = 3 THEN 'APPROVED'
    WHEN status = 4 THEN 'REJECTED'
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
ADD COLUMN status_temp INT;

UPDATE credit_applications 
SET status_temp = CASE 
    WHEN status = 'DRAFT' THEN 0
    WHEN status = 'SCORING' THEN 1
    WHEN status = 'EMPLOYMENT_CHECK' THEN 2
    WHEN status = 'APPROVED' THEN 3
    WHEN status = 'REJECTED' THEN 4
    ELSE 0
END;

ALTER TABLE credit_applications 
DROP COLUMN status;

ALTER TABLE credit_applications 
RENAME COLUMN status_temp TO status;
-- +goose StatementEnd
