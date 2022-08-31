create table `users`
(
    `id`       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ユーザID',
    `name`     VARCHAR(20)     NOT NULL COMMENT 'ユーザ名',
    `password` VARCHAR(80)     NOT NULL COMMENT 'パスワードハッシュ',
    `role`     VARCHAR(80)     NOT NULL COMMENT 'ロール',
    `created`  DATETIME(6)     NOT NULL COMMENT '作成日時',
    `modified` DATETIME(6)     NOT NULL COMMENT '更新日時',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uix_name` (`name`) USING BTREE
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='ユーザ';

create table `tasks`
(
    `id`       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'タスクID',
    `user_id`  BIGINT UNSIGNED NOT NULL COMMENT 'タスクを作成したユーザID',
    `title`    VARCHAR(128)    NOT NULL COMMENT 'タイトル',
    `status`   VARCHAR(20)     NOT NULL COMMENT 'ステータス',
    `created`  DATETIME(6)     NOT NULL COMMENT '作成日時',
    `modified` DATETIME(6)     NOT NULL COMMENT '更新日時',
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
            ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='タスク';
