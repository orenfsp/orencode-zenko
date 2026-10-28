DROP TRIGGER IF EXISTS trg_judge_assignment_notification ON competition_judge;
DROP FUNCTION IF EXISTS notify_judge_assignment();

DROP TRIGGER IF EXISTS trg_judge_category_application_notification ON judge_category_application;
DROP FUNCTION IF EXISTS notify_judge_category_application();
