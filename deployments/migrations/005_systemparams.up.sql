CREATE TABLE `systemparams`
(
    `param` varchar(191) NOT NULL,
    `value` TEXT         NOT NULL,
    PRIMARY KEY (`param`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_general_ci;
