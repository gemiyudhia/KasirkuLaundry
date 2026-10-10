CREATE TABLE orders (
    id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_code     VARCHAR(32)  NOT NULL UNIQUE,
    customer_name  VARCHAR(100) NOT NULL,
    customer_phone VARCHAR(20)  NOT NULL,
    status         ENUM('received','processing','ready','completed') NOT NULL DEFAULT 'received',
    total_amount   BIGINT UNSIGNED NOT NULL DEFAULT 0,
    notes          TEXT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_orders_status (status)
);

CREATE TABLE order_items (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id     BIGINT UNSIGNED NOT NULL,
    service_name VARCHAR(100) NOT NULL,
    unit_price   BIGINT UNSIGNED NOT NULL,
    quantity     DECIMAL(8,2) NOT NULL,
    subtotal     BIGINT UNSIGNED NOT NULL,
    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);