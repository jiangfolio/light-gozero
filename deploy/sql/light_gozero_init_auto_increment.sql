-- auto_increment 模式 DDL（与 IdGen.Mode=auto_increment 配套）
CREATE DATABASE IF NOT EXISTS light_gozero DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE light_gozero;

DROP TABLE IF EXISTS sys_oper_log;
DROP TABLE IF EXISTS sys_login_log;
DROP TABLE IF EXISTS sys_password_history;
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
    id          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    parent_id   BIGINT       NOT NULL DEFAULT 0 COMMENT '父组织ID',
    name        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '组织名称',
    code        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '组织编码',
    sort        INT          NOT NULL DEFAULT 0 COMMENT '排序',
    status      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    created_by  BIGINT       NOT NULL DEFAULT 0 COMMENT '创建人ID',
    updated_by  BIGINT       NOT NULL DEFAULT 0 COMMENT '更新人ID',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_org_code (code),
    KEY idx_sys_org_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统组织表';

CREATE TABLE sys_user (
    id                    BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    org_id                BIGINT       NOT NULL DEFAULT 0 COMMENT '所属组织ID',
    username              VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '登录用户名',
    password_hash         VARCHAR(100) NOT NULL DEFAULT '' COMMENT '密码哈希',
    nickname              VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '昵称',
    status                TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    fail_count            INT          NOT NULL DEFAULT 0 COMMENT '连续登录失败次数',
    locked_until          DATETIME(3)  NULL DEFAULT NULL COMMENT '锁定截止时间',
    password_changed_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '最近改密时间',
    must_change_password  TINYINT      NOT NULL DEFAULT 0 COMMENT '是否强制改密：1是 0否',
    created_by            BIGINT       NOT NULL DEFAULT 0 COMMENT '创建人ID',
    updated_by            BIGINT       NOT NULL DEFAULT 0 COMMENT '更新人ID',
    created_at            DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at            DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_user_username (username),
    KEY idx_sys_user_org (org_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统用户表';

CREATE TABLE sys_role (
    id         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    name       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '角色名称',
    code       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '角色编码',
    data_scope TINYINT      NOT NULL DEFAULT 5 COMMENT '数据范围：1本人 2本部门 3本部门及以下 4指定部门 5全局',
    remark     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    status     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    created_by BIGINT       NOT NULL DEFAULT 0 COMMENT '创建人ID',
    updated_by BIGINT       NOT NULL DEFAULT 0 COMMENT '更新人ID',
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_role_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统角色表';

CREATE TABLE sys_resources (
    id             BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    parent_id      BIGINT       NOT NULL DEFAULT 0 COMMENT '父资源ID',
    name           VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '资源名称/侧栏标题',
    type           VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '类型：directory/menu/button',
    perm_code      VARCHAR(128) NULL DEFAULT NULL COMMENT '权限码',
    path           VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Vue路由path',
    component_path VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Vue组件路径',
    redirect       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '目录重定向',
    route_name     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT 'Vue Router name',
    icon           VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '菜单图标',
    hidden         TINYINT      NOT NULL DEFAULT 0 COMMENT '是否隐藏',
    keep_alive     TINYINT      NOT NULL DEFAULT 0 COMMENT '是否缓存',
    affix          TINYINT      NOT NULL DEFAULT 0 COMMENT '是否固钉',
    sort           INT          NOT NULL DEFAULT 0 COMMENT '排序',
    status         TINYINT      NOT NULL DEFAULT 1 COMMENT '状态',
    created_by     BIGINT       NOT NULL DEFAULT 0 COMMENT '创建人ID',
    updated_by     BIGINT       NOT NULL DEFAULT 0 COMMENT '更新人ID',
    created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_resources_perm_code (perm_code),
    KEY idx_sys_resources_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统资源表';

CREATE TABLE rel_user_role (
    id         BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    user_id    BIGINT NOT NULL COMMENT '用户ID',
    role_id    BIGINT NOT NULL COMMENT '角色ID',
    created_by BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_rel_user_role (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

CREATE TABLE rel_role_resources (
    id            BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    role_id       BIGINT       NOT NULL COMMENT '角色ID',
    resource_id   BIGINT       NOT NULL COMMENT '资源ID',
    resource_type VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '资源类型：directory/menu/button，后续可扩展业务类型',
    created_by    BIGINT       NOT NULL DEFAULT 0 COMMENT '创建人ID',
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_rel_role_resources (role_id, resource_id),
    KEY idx_rel_role_res_type (role_id, resource_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色资源关联表';

CREATE TABLE rel_role_orgs (
    id         BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    role_id    BIGINT NOT NULL COMMENT '角色ID',
    org_id     BIGINT NOT NULL COMMENT '组织ID',
    created_by BIGINT NOT NULL DEFAULT 0 COMMENT '创建人ID',
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_rel_role_orgs (role_id, org_id),
    KEY idx_rel_role_orgs_org (org_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色指定部门关联表';

CREATE TABLE sys_config (
    id           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    config_key   VARCHAR(128) NOT NULL COMMENT '配置键',
    config_value VARCHAR(512) NOT NULL DEFAULT '' COMMENT '配置值',
    value_type   VARCHAR(16)  NOT NULL DEFAULT 'string' COMMENT '值类型：string/int/bool/json',
    group_name   VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '分组：auth/captcha',
    remark       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    status       TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    created_by   BIGINT       NOT NULL DEFAULT 0 COMMENT '创建人ID',
    updated_by   BIGINT       NOT NULL DEFAULT 0 COMMENT '更新人ID',
    created_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_config_key (config_key),
    KEY idx_sys_config_group (group_name, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统动态配置表';

CREATE TABLE sys_dict (
    id         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    parent_id  BIGINT       NOT NULL DEFAULT 0 COMMENT '父ID：0为字典类型，否则为类型ID',
    dict_code  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '字典编码；类型唯一，项冗余父编码',
    label      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '类型名称或项标签',
    value      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '类型固定空串；项取值',
    sort       INT          NOT NULL DEFAULT 0 COMMENT '排序',
    status     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态：1启用 0停用',
    remark     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    created_by BIGINT       NOT NULL DEFAULT 0 COMMENT '创建人ID',
    updated_by BIGINT       NOT NULL DEFAULT 0 COMMENT '更新人ID',
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_sys_dict_code_value (dict_code, value),
    KEY idx_sys_dict_parent (parent_id),
    KEY idx_sys_dict_code (dict_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统字典（类型与项同表）';
