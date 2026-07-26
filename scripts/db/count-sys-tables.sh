#!/bin/bash
for t in sys_config sys_dept sys_dict_data sys_dict_type sys_job sys_job_log sys_logininfor sys_menu sys_notice sys_oper_log sys_post sys_role sys_role_dept sys_role_menu sys_user sys_user_post sys_user_role; do
  c=$(sudo mysql element -N -e "SELECT COUNT(*) FROM $t" 2>/dev/null)
  echo "$t: $c"
done
