CREATE TABLE report_deadline_reminder (
    deadline_id bigint NOT NULL REFERENCES report_deadline (id) ON DELETE CASCADE,
    recipient_id bigint NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    days_before integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (deadline_id, recipient_id, days_before),
    CONSTRAINT report_deadline_reminder_days_check CHECK (days_before > 0)
);
