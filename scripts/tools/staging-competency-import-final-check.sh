#!/usr/bin/env bash
set -euo pipefail

sudo -n mysql element -Nse "
SELECT CONCAT(
  COUNT(*),'|',COUNT(DISTINCT question_code),'|',COUNT(DISTINCT dimension_id),'|',
  SUM(scoring_direction='forward'),'|',SUM(scoring_direction='reverse'),'|',
  SUM(question_status=0),'|',SUM(remark='AI测试题-未信效度验证')
) FROM el_qu WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$';
SELECT CONCAT(COUNT(*),'|',MIN(c),'|',MAX(c),'|',SUM(c)) FROM (
  SELECT dimension_id,COUNT(*) c FROM el_qu
  WHERE question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$'
  GROUP BY dimension_id
) x;
SELECT COUNT(*) FROM el_qu_answer qa
INNER JOIN el_qu q ON q.id=qa.qu_id
WHERE q.question_code REGEXP '^D[0-9]{2}-Q[0-9]{2}$';
"
