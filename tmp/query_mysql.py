import pymysql

conn = pymysql.connect(
    host="127.0.0.1",
    user="root",
    password="root123456",
    port=3306,
    charset="utf8mb4",
)

with conn.cursor() as cur:
    cur.execute(
        """
        SELECT g.id, g.name, i.stock
        FROM shop_good_srv.good g
        JOIN shop_inventory_srv.inventory i ON g.id = i.good
        WHERE i.stock > 0
          AND g.id NOT IN (SELECT good FROM shop_order_srv.shoppingcart)
        ORDER BY i.stock DESC, g.id ASC
        LIMIT 5
        """
    )
    goods = cur.fetchall()
    print("-- goods with stock and not in shoppingcart --")
    for r in goods:
        print(r)

conn.close()
