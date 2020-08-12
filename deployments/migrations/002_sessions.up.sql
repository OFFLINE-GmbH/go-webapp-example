CREATE TABLE IF NOT EXISTS sessions
(
    token  CHAR(43),
    data   BLOB         NOT NULL,
    expiry TIMESTAMP(6) NOT NULL,
    PRIMARY KEY (token)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
