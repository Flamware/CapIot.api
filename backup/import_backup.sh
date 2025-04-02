#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "admin" --dbname "admin" < capiot_backup.sql