CREATE DATABASE IF NOT EXISTS light_gozero DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE light_gozero;

DROP TABLE IF EXISTS sys_oper_log;
DROP TABLE IF EXISTS sys_login_log;
DROP TABLE IF EXISTS sys_password_history;
DROP TABLE IF EXISTS sys_user_shortcut;
DROP TABLE IF EXISTS sys_dict;
DROP TABLE IF EXISTS sys_config;
DROP TABLE IF EXISTS rel_role_orgs;
DROP TABLE IF EXISTS rel_role_resources;
DROP TABLE IF EXISTS rel_user_role;
DROP TABLE IF EXISTS sys_resources;
DROP TABLE IF EXISTS sys_role;
DROP TABLE IF EXISTS sys_user;
DROP TABLE IF EXISTS sys_org;

CREATE TABLE sys_org (
    id          VARCHAR(36)  NOT NULL COMMENT '主键ID',
    parent_id   VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '父组织ID',
    name        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '组织名称',
    code        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '组织编码',
    sort        INT          NOT NULL DEFAULT 0 COMMENT '排序',
    status      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    created_by  VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    updated_by  VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '更新人ID',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_org_code (code),
    KEY idx_sys_org_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统组织表';

CREATE TABLE sys_user (
    id                    VARCHAR(36)  NOT NULL COMMENT '主键ID',
    org_id                VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '所属组织ID',
    username              VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录用户名',
    password_hash         VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码哈希',
    nickname              VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '昵称',
    status                TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    fail_count            INT          NOT NULL DEFAULT 0 COMMENT '连续登录失败次数',
    locked_until          DATETIME(3)  NULL DEFAULT NULL COMMENT '锁定截止时间',
    password_changed_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '最近改密时间',
    must_change_password  TINYINT      NOT NULL DEFAULT 0 COMMENT '是否强制改密：1是 0否',
    created_by            VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    updated_by            VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '更新人ID',
    created_at            DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at            DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_user_username (username),
    KEY idx_sys_user_org (org_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户表';

CREATE TABLE sys_role (
    id         VARCHAR(36)  NOT NULL COMMENT '主键ID',
    name       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '角色名称',
    code       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '角色编码',
    data_scope TINYINT      NOT NULL DEFAULT 5 COMMENT '数据范围：1本人 2本部门 3本部门及以下 4指定部门 5全局',
    remark     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    status     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    created_by VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    updated_by VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '更新人ID',
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_role_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统角色表';

CREATE TABLE sys_resources (
    id             VARCHAR(36)  NOT NULL COMMENT '主键ID',
    parent_id      VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '父资源ID',
    name           VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '资源名称/侧栏标题',
    type           VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '类型：directory/menu/button',
    perm_code      VARCHAR(128) NULL DEFAULT NULL COMMENT '权限码，按钮必填，三段式 module:resource:action',
    path           VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Vue路由path',
    component_path VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Vue组件路径',
    redirect       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '目录重定向',
    route_name     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'Vue Router name',
    icon           VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '菜单图标',
    hidden         TINYINT      NOT NULL DEFAULT 0 COMMENT '是否隐藏：1是 0否',
    keep_alive     TINYINT      NOT NULL DEFAULT 0 COMMENT '是否缓存：1是 0否',
    affix          TINYINT      NOT NULL DEFAULT 0 COMMENT '是否固钉：1是 0否',
    sort           INT          NOT NULL DEFAULT 0 COMMENT '排序',
    status         TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    created_by     VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    updated_by     VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '更新人ID',
    created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_resources_perm_code (perm_code),
    KEY idx_sys_resources_parent (parent_id),
    KEY idx_sys_resources_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统资源表（目录/菜单/按钮）';

CREATE TABLE rel_user_role (
    id         VARCHAR(36) NOT NULL COMMENT '主键ID',
    user_id    VARCHAR(36) NOT NULL COMMENT '用户ID',
    role_id    VARCHAR(36) NOT NULL COMMENT '角色ID',
    created_by VARCHAR(36) NOT NULL DEFAULT '' COMMENT '创建人ID',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_rel_user_role (user_id, role_id),
    KEY idx_rel_user_role_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

CREATE TABLE rel_role_resources (
    id            VARCHAR(36)  NOT NULL COMMENT '主键ID',
    role_id       VARCHAR(36)  NOT NULL COMMENT '角色ID',
    resource_id   VARCHAR(36)  NOT NULL COMMENT '资源ID',
    resource_type VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '资源类型：directory/menu/button，后续可扩展业务类型',
    created_by    VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_rel_role_resources (role_id, resource_id),
    KEY idx_rel_role_resources_res (resource_id),
    KEY idx_rel_role_res_type (role_id, resource_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色资源关联表';

CREATE TABLE rel_role_orgs (
    id         VARCHAR(36) NOT NULL COMMENT '主键ID',
    role_id    VARCHAR(36) NOT NULL COMMENT '角色ID',
    org_id     VARCHAR(36) NOT NULL COMMENT '组织ID',
    created_by VARCHAR(36) NOT NULL DEFAULT '' COMMENT '创建人ID',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_rel_role_orgs (role_id, org_id),
    KEY idx_rel_role_orgs_org (org_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色指定部门关联表';

CREATE TABLE sys_config (
    id           VARCHAR(36)  NOT NULL COMMENT '主键ID',
    config_key   VARCHAR(128) NOT NULL COMMENT '配置键',
    config_value VARCHAR(512) NOT NULL DEFAULT '' COMMENT '配置值',
    value_type   VARCHAR(16)  NOT NULL DEFAULT 'string' COMMENT '值类型：string/int/bool/json',
    group_name   VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '分组：auth/captcha',
    remark       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    status       TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    created_by   VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    updated_by   VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '更新人ID',
    created_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_config_key (config_key),
    KEY idx_sys_config_group (group_name, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统动态配置表';

CREATE TABLE sys_dict (
    id         VARCHAR(36)  NOT NULL COMMENT '主键ID',
    parent_id  VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '父ID：空为字典类型，否则为类型ID',
    dict_code  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '字典编码；类型唯一，项冗余父编码',
    label      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '类型名称或项标签',
    value      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '类型固定空串；项取值',
    sort       INT          NOT NULL DEFAULT 0 COMMENT '排序',
    status     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    remark     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    created_by VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    updated_by VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '更新人ID',
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_dict_code_value (dict_code, value),
    KEY idx_sys_dict_parent (parent_id),
    KEY idx_sys_dict_code (dict_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统字典（类型与项同表）';

CREATE TABLE sys_user_shortcut (
    id          VARCHAR(36)  NOT NULL COMMENT '主键ID',
    user_id     VARCHAR(36)  NOT NULL COMMENT '所属用户ID',
    resource_id VARCHAR(36)  NOT NULL COMMENT '菜单资源ID（名称/路径/图标跟资源走）',
    sort        INT          NOT NULL DEFAULT 0 COMMENT '排序',
    created_by  VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '创建人ID',
    updated_by  VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '更新人ID',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_user_shortcut_resource (user_id, resource_id),
    KEY idx_user_shortcut_user_sort (user_id, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户工作台快捷入口';

CREATE TABLE sys_password_history (
    id            VARCHAR(36)  NOT NULL COMMENT '主键ID',
    user_id       VARCHAR(36)  NOT NULL COMMENT '用户ID',
    password_hash VARCHAR(100) NOT NULL COMMENT '历史密码哈希',
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    KEY idx_pwd_hist_user (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='密码历史表';

CREATE TABLE sys_login_log (
    id         VARCHAR(36)  NOT NULL COMMENT '主键ID',
    user_id    VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '用户ID，失败时可能为空',
    username   VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录用户名',
    ip         VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '客户端IP',
    user_agent VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'User-Agent',
    result     VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '结果：success/fail/locked/logout',
    message    VARCHAR(255) NOT NULL DEFAULT '' COMMENT '说明',
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    KEY idx_login_log_time (created_at),
    KEY idx_login_log_user (username, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录审计日志';

CREATE TABLE sys_oper_log (
    id          VARCHAR(36)  NOT NULL COMMENT '主键ID',
    user_id     VARCHAR(36)  NOT NULL DEFAULT '' COMMENT '操作人ID',
    username    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '操作人用户名',
    module      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '模块',
    action      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '动作',
    method      VARCHAR(16)  NOT NULL DEFAULT '' COMMENT 'HTTP方法',
    path        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请求路径',
    req_body    VARCHAR(2000) NOT NULL DEFAULT '' COMMENT '请求摘要（已脱敏）',
    ip          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '客户端IP',
    latency_ms  INT          NOT NULL DEFAULT 0 COMMENT '耗时毫秒',
    result      VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '结果：success/fail',
    message     VARCHAR(512) NOT NULL DEFAULT '' COMMENT '说明（可读摘要）',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    KEY idx_oper_log_time (created_at),
    KEY idx_oper_log_user (user_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作审计日志';

SET @org_id  = '01a20d49-091b-7f18-b8dd-6018e8e0b384';
SET @user_id = '01a20d49-091b-7f26-8236-e46ad26cd830';
SET @role_id = '01a20d49-091b-7f2c-ad6c-4d9d2222b2e7';
SET @sys     = '';

SET @dir_system   = '01a20d49-091b-7f33-afd9-74ff62926f39';
SET @dir_security = '01a20d49-091d-7001-8000-00000000002c';
SET @menu_security_config = '01a20d49-091d-7001-8000-00000000002d';
SET @btn_security_config_p = '01a20d49-091d-7001-8000-00000000002e';
SET @btn_security_config_u = '01a20d49-091d-7001-8000-00000000002f';
SET @menu_org   = '01a20d49-091b-7f3a-9157-8a8c86989648';
SET @menu_user  = '01a20d49-091c-7002-9aa7-9ed1f5242eef';
SET @menu_role  = '01a20d49-091c-7009-9d34-5addbf69039b';
SET @menu_res   = '01a20d49-091c-700f-9b5c-d7f7358dfff2';

SET @btn_org_c  = '01a20d49-091c-7015-9a59-162ae3267008';
SET @btn_org_u  = '01a20d49-091c-701b-824a-9b1ad60fa564';
SET @btn_org_d  = '01a20d49-091c-7021-bac7-c24bf6dbbfce';
SET @btn_org_dt = '01a20d49-091c-702a-95f2-7c819f922b76';
SET @btn_org_p  = '01a20d49-091c-7030-ba04-676a5504132b';
SET @btn_org_t  = '01a20d49-091d-7001-8000-00000000002a';

SET @btn_user_c  = '01a20d49-091c-7037-8037-a316dfb97eaf';
SET @btn_user_u  = '01a20d49-091c-703c-a1ef-fdac03682c63';
SET @btn_user_d  = '01a20d49-091c-7042-8ae1-443e3a9247c4';
SET @btn_user_dt = '01a20d49-091c-704b-904d-094c03a4d9d0';
SET @btn_user_p  = '01a20d49-091c-7053-91b2-75be012130b6';
SET @btn_user_br = '01a20d49-091c-7059-a856-2102b77dd2dc';
SET @btn_user_ur = '01a20d49-091c-705f-97a1-60b9ba625d06';
SET @btn_user_rs = '01a20d49-091c-7065-bce6-b0ae59289e39';

SET @btn_role_c  = '01a20d49-091c-706b-bc8d-706db1c6a215';
SET @btn_role_u  = '01a20d49-091c-7078-9f94-60657aeb61cc';
SET @btn_role_d  = '01a20d49-091c-7080-aca1-f8a1178bc0af';
SET @btn_role_dt = '01a20d49-091c-7086-9eab-53110e6f9037';
SET @btn_role_p  = '01a20d49-091d-7001-8000-000000000001';
SET @btn_role_br = '01a20d49-091d-7001-8000-000000000002';
SET @btn_role_ur = '01a20d49-091d-7001-8000-000000000003';

SET @btn_res_c  = '01a20d49-091d-7001-8000-000000000004';
SET @btn_res_u  = '01a20d49-091d-7001-8000-000000000005';
SET @btn_res_d  = '01a20d49-091d-7001-8000-000000000006';
SET @btn_res_dt = '01a20d49-091d-7001-8000-000000000007';
SET @btn_res_p  = '01a20d49-091d-7001-8000-000000000008';
SET @btn_res_l  = '01a20d49-091d-7001-8000-000000000031';
SET @btn_res_t  = '01a20d49-091d-7001-8000-00000000002b';
SET @menu_audit_login = '01a20d49-091d-7001-8000-00000000000c';
SET @btn_audit_login = '01a20d49-091d-7001-8000-00000000000d';
SET @btn_audit_oper = '01a20d49-091d-7001-8000-00000000000e';
SET @menu_audit_oper = '01a20d49-091d-7001-8000-000000000029';
SET @menu_config = '01a20d49-091d-7001-8000-00000000000f';
SET @btn_config_p = '01a20d49-091d-7001-8000-000000000010';
SET @btn_config_dt = '01a20d49-091d-7001-8000-000000000011';
SET @btn_config_u = '01a20d49-091d-7001-8000-000000000012';
SET @menu_dict = '01a20d49-091d-7001-8000-000000000013';
SET @btn_dict_c = '01a20d49-091d-7001-8000-000000000014';
SET @btn_dict_u = '01a20d49-091d-7001-8000-000000000015';
SET @btn_dict_d = '01a20d49-091d-7001-8000-000000000016';
SET @btn_dict_dt = '01a20d49-091d-7001-8000-000000000017';
SET @btn_dict_p = '01a20d49-091d-7001-8000-000000000018';
SET @dict_type_user_status = '01a20d49-091d-7001-8000-000000000019';
SET @dict_item_user_status_0 = '01a20d49-091d-7001-8000-00000000001a';
SET @dict_item_user_status_1 = '01a20d49-091d-7001-8000-00000000001b';
SET @cfg_token_ttl = '01a20d49-091d-7001-8000-000000000020';
SET @cfg_max_fail = '01a20d49-091d-7001-8000-000000000021';
SET @cfg_lock_min = '01a20d49-091d-7001-8000-000000000022';
SET @cfg_pwd_min = '01a20d49-091d-7001-8000-000000000023';
SET @cfg_pwd_expire = '01a20d49-091d-7001-8000-000000000024';
SET @cfg_pwd_hist = '01a20d49-091d-7001-8000-000000000025';
SET @cfg_cap_ttl = '01a20d49-091d-7001-8000-000000000026';
SET @cfg_cap_len = '01a20d49-091d-7001-8000-000000000027';
SET @cfg_pwd_complex = '01a20d49-091d-7001-8000-000000000028';
SET @rel_ur = '01a20d49-091d-7001-8000-00000000000a';

INSERT INTO sys_org (id, parent_id, name, code, sort, status, created_by, updated_by) VALUES
(@org_id, '', '默认组织', 'default', 0, 1, @sys, @sys);

INSERT INTO sys_user (id, org_id, username, password_hash, nickname, status, fail_count, locked_until, password_changed_at, must_change_password, created_by, updated_by) VALUES
(@user_id, @org_id, 'superadmin', '$2a$10$h82iSyCZiOwJV9EbIq5awOd3zdB6WGMR8MHfgtzQoBTIpxymvz6zK', '超级管理员', 1, 0, NULL, NOW(3), 0, @sys, @sys);

INSERT INTO sys_role (id, name, code, data_scope, remark, status, created_by, updated_by) VALUES
(@role_id, '超级管理员', 'super_admin', 5, '拥有全部权限', 1, @sys, @sys);

INSERT INTO sys_resources
(id, parent_id, name, type, perm_code, path, component_path, redirect, route_name, icon, hidden, keep_alive, affix, sort, status, created_by, updated_by)
VALUES
(@dir_system, '', '系统管理', 'directory', NULL, '/system', '', '/system/org', 'System', 'setting-two', 0, 0, 0, 1, 1, @sys, @sys),
(@dir_security, '', '网络安全', 'directory', NULL, '/security', '', '/security/config', 'Security', 'protect', 0, 0, 0, 2, 1, @sys, @sys),
(@menu_org, @dir_system, '组织管理', 'menu', NULL, 'org', 'org/index', '', 'SystemOrg', 'building-one', 0, 1, 0, 1, 1, @sys, @sys),
(@menu_user, @dir_system, '用户管理', 'menu', NULL, 'user', 'user/index', '', 'SystemUser', 'user', 0, 1, 0, 2, 1, @sys, @sys),
(@menu_role, @dir_system, '角色管理', 'menu', NULL, 'role', 'role/index', '', 'SystemRole', 'peoples', 0, 1, 0, 3, 1, @sys, @sys),
(@menu_res, @dir_system, '资源管理', 'menu', NULL, 'resource', 'resource/index', '', 'SystemResource', 'building-one', 0, 1, 0, 4, 1, @sys, @sys),
(@menu_config, @dir_system, '系统配置', 'menu', NULL, 'config', 'config/index', '', 'SystemConfig', 'tool', 0, 1, 0, 5, 1, @sys, @sys),
(@menu_dict, @dir_system, '字典管理', 'menu', NULL, 'dict', 'dict/index', '', 'SystemDict', 'book-open', 0, 1, 0, 6, 1, @sys, @sys),
(@menu_security_config, @dir_security, '安全配置', 'menu', NULL, 'config', 'security/config/index', '', 'SecurityConfig', 'protect', 0, 1, 0, 1, 1, @sys, @sys),
(@menu_audit_login, @dir_security, '登录日志', 'menu', NULL, 'login-log', 'audit/login/index', '', 'SecurityLoginLog', 'file-search', 0, 1, 0, 2, 1, @sys, @sys),
(@menu_audit_oper, @dir_security, '操作日志', 'menu', NULL, 'oper-log', 'audit/oper/index', '', 'SecurityOperLog', 'file-search', 0, 1, 0, 3, 1, @sys, @sys),
(@btn_security_config_p, @menu_security_config, '安全配置分页', 'button', 'system:security_config:page', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_security_config_u, @menu_security_config, '安全配置更新', 'button', 'system:security_config:update', '', '', '', '', '', 1, 0, 0, 2, 1, @sys, @sys),
(@btn_org_c,  @menu_org, '组织创建', 'button', 'system:org:create', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_org_u,  @menu_org, '组织更新', 'button', 'system:org:update', '', '', '', '', '', 1, 0, 0, 2, 1, @sys, @sys),
(@btn_org_d,  @menu_org, '组织删除', 'button', 'system:org:delete', '', '', '', '', '', 1, 0, 0, 3, 1, @sys, @sys),
(@btn_org_dt, @menu_org, '组织详情', 'button', 'system:org:detail', '', '', '', '', '', 1, 0, 0, 4, 1, @sys, @sys),
(@btn_org_p,  @menu_org, '组织分页', 'button', 'system:org:page', '', '', '', '', '', 1, 0, 0, 5, 1, @sys, @sys),
(@btn_org_t,  @menu_org, '组织树', 'button', 'system:org:tree', '', '', '', '', '', 1, 0, 0, 6, 1, @sys, @sys),
(@btn_user_c,  @menu_user, '用户创建', 'button', 'system:user:create', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_user_u,  @menu_user, '用户更新', 'button', 'system:user:update', '', '', '', '', '', 1, 0, 0, 2, 1, @sys, @sys),
(@btn_user_d,  @menu_user, '用户删除', 'button', 'system:user:delete', '', '', '', '', '', 1, 0, 0, 3, 1, @sys, @sys),
(@btn_user_dt, @menu_user, '用户详情', 'button', 'system:user:detail', '', '', '', '', '', 1, 0, 0, 4, 1, @sys, @sys),
(@btn_user_p,  @menu_user, '用户分页', 'button', 'system:user:page', '', '', '', '', '', 1, 0, 0, 5, 1, @sys, @sys),
(@btn_user_br, @menu_user, '用户绑角色', 'button', 'system:user:bind_roles', '', '', '', '', '', 1, 0, 0, 6, 1, @sys, @sys),
(@btn_user_ur, @menu_user, '用户解绑角色', 'button', 'system:user:unbind_roles', '', '', '', '', '', 1, 0, 0, 7, 1, @sys, @sys),
(@btn_user_rs, @menu_user, '用户资源', 'button', 'system:user:resources', '', '', '', '', '', 1, 0, 0, 8, 1, @sys, @sys),
(@btn_role_c,  @menu_role, '角色创建', 'button', 'system:role:create', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_role_u,  @menu_role, '角色更新', 'button', 'system:role:update', '', '', '', '', '', 1, 0, 0, 2, 1, @sys, @sys),
(@btn_role_d,  @menu_role, '角色删除', 'button', 'system:role:delete', '', '', '', '', '', 1, 0, 0, 3, 1, @sys, @sys),
(@btn_role_dt, @menu_role, '角色详情', 'button', 'system:role:detail', '', '', '', '', '', 1, 0, 0, 4, 1, @sys, @sys),
(@btn_role_p,  @menu_role, '角色分页', 'button', 'system:role:page', '', '', '', '', '', 1, 0, 0, 5, 1, @sys, @sys),
(@btn_role_br, @menu_role, '角色绑资源', 'button', 'system:role:bind_resources', '', '', '', '', '', 1, 0, 0, 6, 1, @sys, @sys),
(@btn_role_ur, @menu_role, '角色解绑资源', 'button', 'system:role:unbind_resources', '', '', '', '', '', 1, 0, 0, 7, 1, @sys, @sys),
(@btn_res_c,  @menu_res, '资源创建', 'button', 'system:resource:create', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_res_u,  @menu_res, '资源更新', 'button', 'system:resource:update', '', '', '', '', '', 1, 0, 0, 2, 1, @sys, @sys),
(@btn_res_d,  @menu_res, '资源删除', 'button', 'system:resource:delete', '', '', '', '', '', 1, 0, 0, 3, 1, @sys, @sys),
(@btn_res_dt, @menu_res, '资源详情', 'button', 'system:resource:detail', '', '', '', '', '', 1, 0, 0, 4, 1, @sys, @sys),
(@btn_res_p,  @menu_res, '资源分页', 'button', 'system:resource:page', '', '', '', '', '', 1, 0, 0, 5, 1, @sys, @sys),
(@btn_res_l,  @menu_res, '资源列表', 'button', 'system:resource:list', '', '', '', '', '', 1, 0, 0, 6, 1, @sys, @sys),
(@btn_res_t,  @menu_res, '资源树', 'button', 'system:resource:tree', '', '', '', '', '', 1, 0, 0, 7, 1, @sys, @sys),
(@btn_audit_login, @menu_audit_login, '登录日志分页', 'button', 'system:audit:login_page', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_audit_oper, @menu_audit_oper, '操作日志分页', 'button', 'system:audit:oper_page', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_config_p,  @menu_config, '配置分页', 'button', 'system:config:page', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_config_dt, @menu_config, '配置详情', 'button', 'system:config:detail', '', '', '', '', '', 1, 0, 0, 2, 1, @sys, @sys),
(@btn_config_u,  @menu_config, '配置更新', 'button', 'system:config:update', '', '', '', '', '', 1, 0, 0, 3, 1, @sys, @sys),
(@btn_dict_c,  @menu_dict, '字典创建', 'button', 'system:dict:create', '', '', '', '', '', 1, 0, 0, 1, 1, @sys, @sys),
(@btn_dict_u,  @menu_dict, '字典更新', 'button', 'system:dict:update', '', '', '', '', '', 1, 0, 0, 2, 1, @sys, @sys),
(@btn_dict_d,  @menu_dict, '字典删除', 'button', 'system:dict:delete', '', '', '', '', '', 1, 0, 0, 3, 1, @sys, @sys),
(@btn_dict_dt, @menu_dict, '字典详情', 'button', 'system:dict:detail', '', '', '', '', '', 1, 0, 0, 4, 1, @sys, @sys),
(@btn_dict_p,  @menu_dict, '字典分页', 'button', 'system:dict:page', '', '', '', '', '', 1, 0, 0, 5, 1, @sys, @sys);

INSERT INTO sys_dict (id, parent_id, dict_code, label, value, sort, status, remark, created_by, updated_by) VALUES
(@dict_type_user_status, '', 'sys_user_status', '用户状态', '', 0, 1, '用户启用/停用', @sys, @sys),
(@dict_item_user_status_0, @dict_type_user_status, 'sys_user_status', '停用', '0', 1, 1, '', @sys, @sys),
(@dict_item_user_status_1, @dict_type_user_status, 'sys_user_status', '启用', '1', 2, 1, '', @sys, @sys);

INSERT INTO sys_config (id, config_key, config_value, value_type, group_name, remark, status, created_by, updated_by) VALUES
(@cfg_token_ttl, 'auth.token_ttl_seconds', '86400', 'int', 'auth', '会话 Token 有效期（秒）', 1, @sys, @sys),
(@cfg_max_fail, 'auth.max_fail_attempts', '5', 'int', 'auth', '连续登录失败锁定阈值', 1, @sys, @sys),
(@cfg_lock_min, 'auth.lock_minutes', '30', 'int', 'auth', '账号锁定时长（分钟）', 1, @sys, @sys),
(@cfg_pwd_min, 'auth.password_min_len', '8', 'int', 'auth', '密码最小长度', 1, @sys, @sys),
(@cfg_pwd_complex, 'auth.password_complexity_enabled', 'true', 'bool', 'auth', '是否开启密码复杂度（大小写+数字+特殊字符）', 1, @sys, @sys),
(@cfg_pwd_expire, 'auth.password_expire_days', '90', 'int', 'auth', '密码过期天数，0 表示不过期', 1, @sys, @sys),
(@cfg_pwd_hist, 'auth.password_history_count', '5', 'int', 'auth', '禁止复用的历史密码条数', 1, @sys, @sys),
(@cfg_cap_ttl, 'captcha.ttl_seconds', '120', 'int', 'captcha', '验证码有效期（秒）', 1, @sys, @sys),
(@cfg_cap_len, 'captcha.length', '4', 'int', 'captcha', '验证码位数', 1, @sys, @sys);

INSERT INTO rel_user_role (id, user_id, role_id, created_by) VALUES (@rel_ur, @user_id, @role_id, @sys);

INSERT INTO rel_role_resources (id, role_id, resource_id, resource_type, created_by)
SELECT UUID(), @role_id, id, type, @sys FROM sys_resources;
