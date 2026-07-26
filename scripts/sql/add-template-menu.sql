-- 插入报告模板菜单项（parent_id=2000 是测评管理）
INSERT INTO sys_menu (menu_id, menu_name, parent_id, order_num, path, component, menu_type, visible, status, icon, create_by, create_time, update_time) 
VALUES (2100, '报告模板', 2000, 5, 'template', 'exam/template/index', 'C', '0', '0', 'documentation', 'admin', NOW(), NOW()) 
ON DUPLICATE KEY UPDATE menu_name='报告模板', component='exam/template/index', path='template';
