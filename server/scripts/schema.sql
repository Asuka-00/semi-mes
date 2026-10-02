CREATE TABLE sys_user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL,
  password TEXT NOT NULL,
  real_name TEXT,
  email TEXT,
  phone TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE sys_role (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  role_code TEXT NOT NULL,
  role_name TEXT NOT NULL,
  description TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE sys_menu (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  parent_id INTEGER NOT NULL DEFAULT 0,
  menu_type INTEGER NOT NULL,
  menu_name TEXT NOT NULL,
  permission_code TEXT,
  route_name TEXT,
  route_path TEXT,
  component_path TEXT,
  icon TEXT,
  sort_order INTEGER NOT NULL DEFAULT 0,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE sys_user_role (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  role_id INTEGER NOT NULL,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE sys_role_menu (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  role_id INTEGER NOT NULL,
  menu_id INTEGER NOT NULL,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE base_factory (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  factory_code TEXT NOT NULL,
  factory_name TEXT NOT NULL,
  address TEXT,
  contact TEXT,
  phone TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE base_workshop (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  factory_id INTEGER NOT NULL,
  workshop_code TEXT NOT NULL,
  workshop_name TEXT NOT NULL,
  workshop_type TEXT,
  description TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE base_production_line (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workshop_id INTEGER NOT NULL,
  line_code TEXT NOT NULL,
  line_name TEXT NOT NULL,
  capacity INTEGER NOT NULL DEFAULT 0,
  description TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE base_product (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  product_code TEXT NOT NULL,
  product_name TEXT NOT NULL,
  product_type TEXT,
  version TEXT,
  description TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE base_process_route (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  product_id INTEGER NOT NULL,
  route_code TEXT NOT NULL,
  route_name TEXT NOT NULL,
  version TEXT,
  is_default INTEGER NOT NULL DEFAULT 0,
  description TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE base_operation (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  route_id INTEGER NOT NULL,
  operation_code TEXT NOT NULL,
  operation_name TEXT NOT NULL,
  operation_type TEXT,
  sequence INTEGER NOT NULL,
  standard_time INTEGER NOT NULL DEFAULT 0,
  description TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);

CREATE TABLE base_recipe (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  operation_id INTEGER NOT NULL,
  recipe_code TEXT NOT NULL,
  recipe_name TEXT NOT NULL,
  version TEXT,
  parameters TEXT,
  is_default INTEGER NOT NULL DEFAULT 0,
  description TEXT,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME,
  updated_at DATETIME,
  deleted_at DATETIME
);
