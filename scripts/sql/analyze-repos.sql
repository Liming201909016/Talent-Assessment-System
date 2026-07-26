SELECT r.id, r.code, r.title,
  COUNT(DISTINCT qr.qu_id) AS qu_cnt,
  AVG(opts.opt_cnt) AS avg_opts_per_qu,
  SUM(CASE WHEN opts.right_cnt > 0 THEN 1 ELSE 0 END) AS qu_with_anchor,
  MIN(opts.right_cnt) AS min_right, MAX(opts.right_cnt) AS max_right
FROM el_repo r
JOIN el_qu_repo qr ON qr.repo_id=r.id
JOIN (
  SELECT qu_id, COUNT(*) AS opt_cnt, SUM(is_right) AS right_cnt
  FROM el_qu_answer GROUP BY qu_id
) opts ON opts.qu_id=qr.qu_id
GROUP BY r.id, r.code, r.title
ORDER BY r.code;
