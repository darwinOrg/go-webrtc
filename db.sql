CREATE TABLE `webrtc_room`
(
    `id`          BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `biz_type`    VARCHAR(64) NOT NULL COMMENT '业务类型',
    `biz_id`      BIGINT      NOT NULL COMMENT '业务ID',
    `room_id`     VARCHAR(64) NOT NULL COMMENT '房间ID',
    `room_name`   VARCHAR(64) NOT NULL DEFAULT '' COMMENT '房间名称',
    `room_status` INT         NOT NULL DEFAULT 0 COMMENT '房间状态',
    `created_by`  bigint      NOT NULL DEFAULT '0',
    `created_at`  datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `modified_by` bigint      NOT NULL DEFAULT '0',
    `modified_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_room_id` (`room_id`),
    UNIQUE KEY `uk_biz_type_biz_id` (`biz_type`, `biz_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='webrtc房间';

CREATE TABLE `webrtc_room_client`
(
    `id`            BIGINT      NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `room_id`       VARCHAR(64) NOT NULL COMMENT '房间ID',
    `client_id`     VARCHAR(64) NOT NULL COMMENT '客户ID',
    `client_type`   INT         NOT NULL DEFAULT 0 COMMENT '客户类型',
    `client_status` INT         NOT NULL DEFAULT 0 COMMENT '客户状态',
    `user_id`       bigint      NOT NULL DEFAULT '0' COMMENT '用户ID',
    `created_by`    bigint      NOT NULL DEFAULT '0',
    `created_at`    datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `modified_by`   bigint      NOT NULL DEFAULT '0',
    `modified_at`   datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_room_id_client_id` (`room_id`, `client_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT ='webrtc房间客户';
