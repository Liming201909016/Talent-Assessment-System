#!/usr/bin/env bash
set -euo pipefail
sudo -n mysql element -Nse "
SELECT CONCAT(table_name,'.',column_name,'=',COALESCE(character_set_name,'-'),'/',COALESCE(collation_name,'-'))
FROM information_schema.columns
WHERE table_schema='element'
  AND ((table_name='el_exam' AND column_name='id')
    OR (table_name='el_competency_dimension' AND column_name IN ('id','code','name'))
    OR (table_name='el_exam_competency_dimension' AND column_name IN ('id','exam_id','dimension_id')))
ORDER BY table_name,column_name;
"
