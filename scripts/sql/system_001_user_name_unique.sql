-- sys_user.user_name concurrency-safe uniqueness hardening
-- MySQL 5.7 compatible and idempotent. This script never modifies or removes user rows.
SET NAMES utf8mb4;

SELECT user_name, COUNT(*) AS duplicate_count
FROM sys_user
GROUP BY user_name
HAVING COUNT(*) > 1;

SET @duplicate_user_names := (
    SELECT COUNT(*)
    FROM (
        SELECT user_name
        FROM sys_user
        GROUP BY user_name
        HAVING COUNT(*) > 1
    ) AS duplicate_names
);

SET @duplicate_guard_sql := IF(
    @duplicate_user_names = 0,
    'SELECT ''sys_user.user_name duplicate check passed'' AS migration_status',
    'SIGNAL SQLSTATE ''45000'' SET MESSAGE_TEXT = ''Duplicate sys_user.user_name values exist; resolve them before adding the unique index'''
);
PREPARE duplicate_guard_stmt FROM @duplicate_guard_sql;
EXECUTE duplicate_guard_stmt;
DEALLOCATE PREPARE duplicate_guard_stmt;

SET @unique_index_exists := (
    SELECT COUNT(*)
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'sys_user'
      AND index_name = 'uk_sys_user_user_name'
      AND non_unique = 0
);

SET @create_unique_index_sql := IF(
    @unique_index_exists = 0,
    'ALTER TABLE sys_user ADD UNIQUE INDEX uk_sys_user_user_name (user_name)',
    'SELECT ''uk_sys_user_user_name already exists'' AS migration_status'
);
PREPARE create_unique_index_stmt FROM @create_unique_index_sql;
EXECUTE create_unique_index_stmt;
DEALLOCATE PREPARE create_unique_index_stmt;
