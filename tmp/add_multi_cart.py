import pymysql

conn = pymysql.connect(host="127.0.0.1", user="root", password="root123456", port=3306, charset="utf8mb4", autocommit=False)
try:
    with conn.cursor() as cur:
        # 1) ensure composite unique index for (user, good)
        cur.execute("SHOW INDEX FROM shop_order_srv.shoppingcart")
        idx_names = {row[2] for row in cur.fetchall()}
        if "idx_user" in idx_names:
            cur.execute("ALTER TABLE shop_order_srv.shoppingcart DROP INDEX idx_user")
        if "idx_good" in idx_names:
            cur.execute("ALTER TABLE shop_order_srv.shoppingcart DROP INDEX idx_good")
        if "idx_user_good" not in idx_names:
            cur.execute("ALTER TABLE shop_order_srv.shoppingcart ADD UNIQUE KEY idx_user_good (`user`, `good`)")

        # 2) pick goods with stock and not already in user=1 cart
        cur.execute(
            """
            SELECT g.id
            FROM shop_good_srv.good g
            JOIN shop_inventory_srv.inventory i ON g.id = i.good
            WHERE i.stock > 0
              AND g.id NOT IN (SELECT good FROM shop_order_srv.shoppingcart WHERE user = 1)
            ORDER BY i.stock DESC, g.id ASC
            LIMIT 3
            """
        )
        goods = [row[0] for row in cur.fetchall()]
        if not goods:
            raise RuntimeError("no available goods with stock to insert into cart")

        # 3) insert into cart for user=1
        for gid in goods:
            cur.execute(
                """
                INSERT INTO shop_order_srv.shoppingcart
                (`user`, `good`, `nums`, `checked`, `add_time`, `update_time`)
                VALUES (1, %s, 1, 1, NOW(3), NOW(3))
                """,
                (gid,)
            )

    conn.commit()
    print(f"OK. inserted goods for user=1 cart: {goods}")
except Exception as e:
    conn.rollback()
    print("ERROR:", e)
    raise
finally:
    conn.close()
