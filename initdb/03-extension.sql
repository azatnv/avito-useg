CREATE EXTENSION pg_cron;

SELECT cron.schedule('0 4 * * *', $$DELETE FROM user2seg WHERE date_end > now()$$);
