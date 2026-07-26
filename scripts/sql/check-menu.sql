SELECT menu_id, menu_name, parent_id, path, component, menu_type FROM sys_menu WHERE menu_name LIKE '%测评%' OR menu_name LIKE '%报告%' OR menu_id IN (2000, 2100) ORDER BY parent_id, order_num;
