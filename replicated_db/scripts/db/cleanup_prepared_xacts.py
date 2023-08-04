#!/usr/bin/env python
#
# Purges all prepared xacts from the specified database
# creadits: https://stackoverflow.com/questions/25077581/how-to-rollback-all-open-postgresql-transactions

import sys
import psycopg2
import subprocess


if len(sys.argv) != 2:
    print('Usage: cleanup_prepared_xacts.py "dbname=mydb ..."...')

n = len(sys.argv)

for i in range(1, n):
    conn = psycopg2.connect(sys.argv[i])
    conn.set_isolation_level(psycopg2.extensions.ISOLATION_LEVEL_AUTOCOMMIT)

    curs = conn.cursor()
    curs.execute("SELECT gid FROM pg_prepared_xacts WHERE database = current_database()")
    for (gid,) in curs.fetchall():
        curs.execute("ROLLBACK PREPARED %s", (gid,))