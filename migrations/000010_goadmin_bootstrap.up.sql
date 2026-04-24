-- ============================================================
-- Migration 010: Bootstrap GoAdmin Internal Schema
-- ============================================================

-- Core GoAdmin tables
CREATE TABLE IF NOT EXISTS goadmin_menu (
    id          SERIAL PRIMARY KEY,
    parent_id   INT NOT NULL DEFAULT 0,
    type        INT DEFAULT 0,
    "order"     INT NOT NULL DEFAULT 0,
    title       VARCHAR(50) NOT NULL,
    header      VARCHAR(100),
    plugin_name VARCHAR(100) NOT NULL DEFAULT '',
    icon        VARCHAR(50) NOT NULL DEFAULT '',
    uri         VARCHAR(3000) NOT NULL,
    uuid        VARCHAR(100),
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_operation_log (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL,
    path        VARCHAR(255) NOT NULL,
    method      VARCHAR(10) NOT NULL,
    ip          VARCHAR(15) NOT NULL,
    input       TEXT NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_site (
    id          SERIAL PRIMARY KEY,
    key         VARCHAR(100) NOT NULL,
    value       TEXT NOT NULL,
    type        INT DEFAULT 0,
    description VARCHAR(3000),
    state       INT DEFAULT 0,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_permissions (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(50) NOT NULL,
    slug         VARCHAR(50) NOT NULL,
    http_method  VARCHAR(255),
    http_path    TEXT NOT NULL,
    created_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_roles (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_users (
    id             SERIAL PRIMARY KEY,
    username       VARCHAR(100) NOT NULL,
    password       VARCHAR(100) NOT NULL,
    name           VARCHAR(100) NOT NULL,
    avatar         VARCHAR(255),
    remember_token VARCHAR(100),
    created_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at     TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_role_menu (
    role_id     INT NOT NULL,
    menu_id     INT NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_role_permissions (
    role_id      INT NOT NULL,
    permission_id INT NOT NULL,
    created_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_role_users (
    role_id     INT NOT NULL,
    user_id     INT NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_user_permissions (
    user_id      INT NOT NULL,
    permission_id INT NOT NULL,
    created_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at   TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS goadmin_session (
    id          SERIAL PRIMARY KEY,
    sid         VARCHAR(50) NOT NULL,
    "values"    VARCHAR(3000) NOT NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now(),
    updated_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT now()
);

-- Idempotent uniqueness guards
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_users_username ON goadmin_users(username);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_roles_slug ON goadmin_roles(slug);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_permissions_slug ON goadmin_permissions(slug);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_site_key ON goadmin_site(key);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_session_sid ON goadmin_session(sid);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_role_menu_pair ON goadmin_role_menu(role_id, menu_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_role_permissions_pair ON goadmin_role_permissions(role_id, permission_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_role_users_pair ON goadmin_role_users(role_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_goadmin_user_permissions_pair ON goadmin_user_permissions(user_id, permission_id);

-- Default users
INSERT INTO goadmin_users (username, password, name, avatar, remember_token, created_at, updated_at)
SELECT
    'admin',
    '$2a$10$OxWYJJGTP2gi00l2x06QuOWqw5VR47MQCJ0vNKnbMYfrutij10Hwe',
    'admin',
    '',
    NULL,
    now(),
    now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_users WHERE username = 'admin'
);

INSERT INTO goadmin_users (username, password, name, avatar, remember_token, created_at, updated_at)
SELECT
    'operator',
    '$2a$10$rVqkOzHjN2MdlEprRflb1eGP0oZXuSrbJLOmJagFsCd81YZm0bsh.',
    'Operator',
    '',
    NULL,
    now(),
    now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_users WHERE username = 'operator'
);

-- Default roles and permissions
INSERT INTO goadmin_roles (name, slug, created_at, updated_at)
SELECT 'Administrator', 'administrator', now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_roles WHERE slug = 'administrator'
);

INSERT INTO goadmin_roles (name, slug, created_at, updated_at)
SELECT 'Operator', 'operator', now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_roles WHERE slug = 'operator'
);

INSERT INTO goadmin_permissions (name, slug, http_method, http_path, created_at, updated_at)
SELECT 'All permission', '*', '', '*', now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_permissions WHERE slug = '*'
);

INSERT INTO goadmin_permissions (name, slug, http_method, http_path, created_at, updated_at)
SELECT 'Dashboard', 'dashboard', 'GET,PUT,POST,DELETE', '/', now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_permissions WHERE slug = 'dashboard'
);

-- Default menu tree
INSERT INTO goadmin_menu (parent_id, type, "order", title, header, plugin_name, icon, uri, created_at, updated_at)
SELECT 0, 1, 1, 'Dashboard', NULL, '', 'fa-bar-chart', '/', now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_menu WHERE title = 'Dashboard' AND uri = '/'
);

INSERT INTO goadmin_menu (parent_id, type, "order", title, header, plugin_name, icon, uri, created_at, updated_at)
SELECT 0, 1, 2, 'Admin', NULL, '', 'fa-tasks', '', now(), now()
WHERE NOT EXISTS (
    SELECT 1 FROM goadmin_menu WHERE title = 'Admin' AND uri = ''
);

INSERT INTO goadmin_menu (parent_id, type, "order", title, header, plugin_name, icon, uri, created_at, updated_at)
SELECT m.id, 1, 2, 'Users', NULL, '', 'fa-users', '/info/manager', now(), now()
FROM goadmin_menu m
WHERE m.title = 'Admin' AND m.uri = ''
  AND NOT EXISTS (
      SELECT 1 FROM goadmin_menu WHERE title = 'Users' AND uri = '/info/manager'
  )
LIMIT 1;

INSERT INTO goadmin_menu (parent_id, type, "order", title, header, plugin_name, icon, uri, created_at, updated_at)
SELECT m.id, 1, 3, 'Roles', NULL, '', 'fa-user', '/info/roles', now(), now()
FROM goadmin_menu m
WHERE m.title = 'Admin' AND m.uri = ''
  AND NOT EXISTS (
      SELECT 1 FROM goadmin_menu WHERE title = 'Roles' AND uri = '/info/roles'
  )
LIMIT 1;

INSERT INTO goadmin_menu (parent_id, type, "order", title, header, plugin_name, icon, uri, created_at, updated_at)
SELECT m.id, 1, 4, 'Permission', NULL, '', 'fa-ban', '/info/permission', now(), now()
FROM goadmin_menu m
WHERE m.title = 'Admin' AND m.uri = ''
  AND NOT EXISTS (
      SELECT 1 FROM goadmin_menu WHERE title = 'Permission' AND uri = '/info/permission'
  )
LIMIT 1;

INSERT INTO goadmin_menu (parent_id, type, "order", title, header, plugin_name, icon, uri, created_at, updated_at)
SELECT m.id, 1, 5, 'Menu', NULL, '', 'fa-bars', '/menu', now(), now()
FROM goadmin_menu m
WHERE m.title = 'Admin' AND m.uri = ''
  AND NOT EXISTS (
      SELECT 1 FROM goadmin_menu WHERE title = 'Menu' AND uri = '/menu'
  )
LIMIT 1;

INSERT INTO goadmin_menu (parent_id, type, "order", title, header, plugin_name, icon, uri, created_at, updated_at)
SELECT m.id, 1, 6, 'Operation log', NULL, '', 'fa-history', '/info/op', now(), now()
FROM goadmin_menu m
WHERE m.title = 'Admin' AND m.uri = ''
  AND NOT EXISTS (
      SELECT 1 FROM goadmin_menu WHERE title = 'Operation log' AND uri = '/info/op'
  )
LIMIT 1;

-- Role mappings
INSERT INTO goadmin_role_users (role_id, user_id, created_at, updated_at)
SELECT r.id, u.id, now(), now()
FROM goadmin_roles r
JOIN goadmin_users u ON u.username = 'admin'
WHERE r.slug = 'administrator'
  AND NOT EXISTS (
      SELECT 1
      FROM goadmin_role_users ru
      WHERE ru.role_id = r.id AND ru.user_id = u.id
  );

INSERT INTO goadmin_role_users (role_id, user_id, created_at, updated_at)
SELECT r.id, u.id, now(), now()
FROM goadmin_roles r
JOIN goadmin_users u ON u.username = 'operator'
WHERE r.slug = 'operator'
  AND NOT EXISTS (
      SELECT 1
      FROM goadmin_role_users ru
      WHERE ru.role_id = r.id AND ru.user_id = u.id
  );

INSERT INTO goadmin_role_permissions (role_id, permission_id, created_at, updated_at)
SELECT r.id, p.id, now(), now()
FROM goadmin_roles r
JOIN goadmin_permissions p ON p.slug = '*'
WHERE r.slug = 'administrator'
  AND NOT EXISTS (
      SELECT 1
      FROM goadmin_role_permissions rp
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

INSERT INTO goadmin_role_permissions (role_id, permission_id, created_at, updated_at)
SELECT r.id, p.id, now(), now()
FROM goadmin_roles r
JOIN goadmin_permissions p ON p.slug = 'dashboard'
WHERE r.slug IN ('administrator', 'operator')
  AND NOT EXISTS (
      SELECT 1
      FROM goadmin_role_permissions rp
      WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

INSERT INTO goadmin_user_permissions (user_id, permission_id, created_at, updated_at)
SELECT u.id, p.id, now(), now()
FROM goadmin_users u
JOIN goadmin_permissions p ON p.slug = '*'
WHERE u.username = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM goadmin_user_permissions up
      WHERE up.user_id = u.id AND up.permission_id = p.id
  );

-- Admin role can access Admin + Dashboard menu nodes
INSERT INTO goadmin_role_menu (role_id, menu_id, created_at, updated_at)
SELECT r.id, m.id, now(), now()
FROM goadmin_roles r
JOIN goadmin_menu m ON m.title = 'Admin' AND m.uri = ''
WHERE r.slug = 'administrator'
  AND NOT EXISTS (
      SELECT 1
      FROM goadmin_role_menu rm
      WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

INSERT INTO goadmin_role_menu (role_id, menu_id, created_at, updated_at)
SELECT r.id, m.id, now(), now()
FROM goadmin_roles r
JOIN goadmin_menu m ON m.title = 'Dashboard' AND m.uri = '/'
WHERE r.slug IN ('administrator', 'operator')
  AND NOT EXISTS (
      SELECT 1
      FROM goadmin_role_menu rm
      WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );
