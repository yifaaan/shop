import pymysql

conn = pymysql.connect(host="127.0.0.1", user="root", password="root123456", port=3306, charset="utf8mb4", autocommit=False)
try:
    with conn.cursor() as cur:
        # 1) insert good
        cur.execute(
            """
            INSERT INTO `shop_good_srv`.`good`
            (`name`, `good_sn`, `category_id`, `brand_id`, `on_sale`, `ship_free`, `is_new`, `is_hot`,
             `click_num`, `sold_num`, `fav_num`, `market_price`, `shop_price`,
             `good_brief`, `good_front_image`, `images`, `desc_images`,
             `add_time`, `update_time`)
            VALUES
            ('test-cart-good', 'TEST-CART-001', 130358, 614, 1, 0, 1, 0,
             0, 0, 0, 99.9, 79.9,
             'test cart good', 'https://example.com/cover.png', '[]', '[]',
             NOW(3), NOW(3))
            """
        )
        good_id = cur.lastrowid

        # 2) insert inventory
        cur.execute(
            """
            INSERT INTO `shop_inventory_srv`.`inventory`
            (`good`, `stock`, `version`, `add_time`, `update_time`)
            VALUES
            (%s, 100, 0, NOW(3), NOW(3))
            """,
            (good_id,)
        )

        # 3) insert shopping cart for user=1
        cur.execute(
            """
            INSERT INTO `shop_order_srv`.`shoppingcart`
            (`user`, `good`, `nums`, `checked`, `add_time`, `update_time`)
            VALUES
            (1, %s, 2, 1, NOW(3), NOW(3))
            """,
            (good_id,)
        )

    conn.commit()
    print(f"OK. inserted good_id={good_id} for user=1 cart")
except Exception as e:
    conn.rollback()
    print("ERROR:", e)
    raise
finally:
    conn.close()
