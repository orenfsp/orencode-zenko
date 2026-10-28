DROP TRIGGER IF EXISTS trg_competition_result_notification ON competition_result;
DROP FUNCTION IF EXISTS notify_competition_result();

DROP TRIGGER IF EXISTS trg_competition_registration_notification ON competition;
DROP FUNCTION IF EXISTS notify_competition_registration_opened();

DROP TRIGGER IF EXISTS trg_rank_application_notification ON rank_application;
DROP FUNCTION IF EXISTS notify_rank_application_created();