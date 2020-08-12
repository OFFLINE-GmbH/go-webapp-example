CREATE TABLE IF NOT EXISTS quotes
(
    id         SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT,
    author     VARCHAR(32),
    content    TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
