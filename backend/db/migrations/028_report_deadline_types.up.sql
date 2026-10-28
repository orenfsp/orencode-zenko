ALTER TABLE report_deadline
    ADD COLUMN deadline_type varchar(32) NOT NULL DEFAULT 'report';

ALTER TABLE report_deadline
    ADD CONSTRAINT report_deadline_type_check
    CHECK (deadline_type IN ('report', 'governor_award'));
