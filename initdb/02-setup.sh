#!/bin/sh

# Cron setup
printf "
shared_preload_libraries='pg_cron'\n
cron.database_name='%s'\n
" "${POSTGRES_DB}" >> "${PGDATA}"/postgresql.conf

printf "
host %s %s localhost trust
" "${POSTGRES_DB}" "${POSTGRES_USER}" >> "${PGDATA}"/pg_hba.conf

# Required to load pg_cron
pg_ctl restart
