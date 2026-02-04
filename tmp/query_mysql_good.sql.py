import pymysql

conn = pymysql.connect(host="127.0.0.1", user="root", password="root123456", port=3306, charset="utf8mb4")
with conn.cursor() as cur:
    cur.execute("SHOW CREATE TABLE shop_good_srv.good")
    row = cur.fetchone()
    print("-- good schema --")
    if row:
        print(row[1])

    cur.execute("SHOW CREATE TABLE shop_inventory_srv.inventory")
    row = cur.fetchone()
    print("-- inventory schema --")
    if row:
        print(row[1])

    cur.execute("SHOW CREATE TABLE shop_order_srv.shoppingcart")
    row = cur.fetchone()
    print("-- shoppingcart schema --")
    if row:
        print(row[1])

conn.close()
