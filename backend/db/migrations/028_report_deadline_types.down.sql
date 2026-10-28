ALTER TABLE report_deadline DROP CONSTRAINT IF EXISTS report_deadline_type_check;
ALTER TABLE report_deadline DROP COLUMN IF EXISTS deadline_type;
