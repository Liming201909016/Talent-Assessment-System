SET NAMES utf8mb4;

BEGIN;

-- Unsupported modules are retained as historical menu rows but cannot be
-- displayed or assigned after their routes and pages are retired.
UPDATE sys_menu
SET status = '1',
    visible = '1',
    update_by = 'system-retirement',
    update_time = NOW()
WHERE menu_id IN (
    1041, 1042,
    1044, 1045,
    109, 1046, 1047, 1048,
    110, 1049, 1050, 1051, 1052, 1053, 1054,
    113,
    115, 1055, 1056, 1057, 1058, 1059, 1060
);

COMMIT;