import pymysql

conn = pymysql.connect(host="127.0.0.1", user="root", password="root123456", port=3306, charset="utf8mb4")
with conn.cursor() as cur:
    cur.execute("SELECT id, name FROM shop_good_srv.category ORDER BY id ASC LIMIT 5")
    print("-- category sample --")
    for r in cur.fetchall():
        print(r)

    cur.execute("SELECT id, name FROM shop_good_srv.brand ORDER BY id ASC LIMIT 5")
    print("-- brand sample --")
    for r in cur.fetchall():
        print(r)

conn.close()
