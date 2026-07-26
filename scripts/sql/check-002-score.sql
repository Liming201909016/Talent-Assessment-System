SELECT eq.content AS V, ea.id, ea.content, ea.is_right, ea.score
FROM el_qu eq
JOIN el_qu_repo qr ON qr.qu_id=eq.id
JOIN el_qu_answer ea ON ea.qu_id=eq.id
WHERE qr.repo_id='2019583026766479361' AND qr.sort=1
ORDER BY ea.id LIMIT 8;
