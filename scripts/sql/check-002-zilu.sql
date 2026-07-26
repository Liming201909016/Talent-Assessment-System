SELECT @pid := paper_id FROM el_candidate WHERE name='888' LIMIT 1;
SELECT @pid;
SELECT eq.content, pq.actual_score
FROM el_paper_qu pq JOIN el_qu eq ON eq.id=pq.qu_id
WHERE pq.paper_id=@pid AND eq.content IN ('V11','V25','V39','V53','V69','V84','V97','V112','V125','V138')
ORDER BY CAST(SUBSTRING(eq.content,2) AS UNSIGNED);
