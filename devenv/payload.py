#!/usr/bin/env python3

import pymysql
import argparse
import time

parser = argparse.ArgumentParser(
    description="Run a dumb MySQL payload in an infinite loop."
)
parser.add_argument("--host", default="127.0.0.1", help="MySQL Host")
parser.add_argument("--user", default="payload", help="MySQL User")
parser.add_argument("--password", required=True, help="MySQL Password")
parser.add_argument("--database", default="test", help="MySQL Database to connect to")
parser.add_argument("--table", required=True, help="MySQL Table to write to")

args = parser.parse_args()


def get_connection():
    return pymysql.connect(
        host=args.host,
        user=args.user,
        password=args.password,
        database=args.database,
        autocommit=True,
    )


while True:
    conn = get_connection()
    cursor = conn.cursor()
    table = args.table

    cursor.execute(
        f"CREATE TABLE IF NOT EXISTS {table} (id INT AUTO_INCREMENT PRIMARY KEY, data VARCHAR(255))"
    )

    print("\n[INFO] Inserting 10 records (one per second)...")
    for i in range(10):
        cursor.execute(
            f"INSERT INTO {table} (data) VALUES (%s)",
            (f"DumbData-{i}",),
        )
        print(f"[INSERT] DumbData-{i}")
        time.sleep(1)

    print("\n[INFO] Deleting all records...")
    cursor.execute(f"DELETE FROM {table}")

    cursor.close()
    conn.close()

    print("[LOOP] Restarting...")
