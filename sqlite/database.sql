/*comment '0: active, 1: inactive, 2: locked */
CREATE TABLE IF NOT EXISTS `user` 
(
    `id`            INTEGER PRIMARY KEY AUTOINCREMENT,
    `username`      VARCHAR(64) NOT NULL,
    `password`      VARCHAR(128),
    `nickname`      VARCHAR(24),
    `status`        INTEGER NOT NULL DEFAULT 0,
    `is_root`       INTEGER NOT NULL default 0,
    `login_at`     INTEGER NOT NULL default 0,
    `created_at`     INTEGER NOT NULL default 0,
    `updated_at`    INTEGER NOT NULL default 0,
    UNIQUE (`username`)
)

INSERT INTO user (`username`,`password`,`nickname`,`status`,`is_root`,`login_at`,`created_at`,`updated_at`)  VALUES("admin","$2a$10$XhL194U090xIMBDob4X9Uu/NMiPYEtGgIsvajLbTZv.BYNG5202sa","",0,1,0,strftime('%s','now'),strftime('%s','now'));

CREATE TABLE IF NOT EXISTS `cloud_provider`
(
    `id`  INTEGER PRIMARY KEY AUTOINCREMENT,
    `name` VARCHAR(24) NOT NULL,
    `account` VARCHAR(48) NOT NULL,
    `type` VARCHAR(24) NOT NULL,
    `api_checked` INTEGER NOT NULL default 0,
    `last_checked_at` INTEGER NOT NULL default 0,
    `api_config`  VARCHAR(128) NOT NULL,
    `created_at`  INTEGER NOT NULL default 0,
    `updated_at`    INTEGER NOT NULL default 0
)

CREATE TABLE IF NOT EXISTS`cloud_provider_ops_logs`
(
    `id`  INTEGER PRIMARY KEY AUTOINCREMENT,
    `cloud_provider_id` INTEGER NOT NULL,
    `api_path`  VARCHAR(256) NOT NULL,
    `api_body` VARCHAR(256) NOT NULL,
    `api_response` VARCHAR(512) NOT NULL,
    `api_result` INTEGER NOT NULL,
    `created_at`  INTEGER NOT NULL default 0
)

CREATE TABLE IF NOT EXISTS `ansible_ops_logs`
(
    `id`  INTEGER PRIMARY KEY AUTOINCREMENT,
    `cloud_provider_id` INTEGER NOT NULL,
    `instance_id` VARCHAR(64) NOT NULL,
    `ansible_playbook` VARCHAR(512) NOT NULL,
    `ansible_host_config` VARCHAR(128) NOT NULL,
    `play_cmd` VARCHAR(512) NOT NULL,
    `play_result` VARCHAR(512) NOT NULL,
    `status`  VARCHAR(32) NOT NULL,
    `created_at`  INTEGER NOT NULL default 0,
    `updated_at`  INTEGER NOT NULL default 0
)