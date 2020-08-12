CREATE TABLE IF NOT EXISTS auditlogs
(
    id          INT UNSIGNED NOT NULL AUTO_INCREMENT,

    user_id     INT UNSIGNED,

    field       varchar(191),
    value_new   text,
    value_old   text,

    action      VARCHAR(191) NOT NULL,
    entity_type VARCHAR(191),
    entity_id   INT UNSIGNED,
    meta        TEXT,

    created_at  TIMESTAMP,
    updated_at  TIMESTAMP,

    PRIMARY KEY (id)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
