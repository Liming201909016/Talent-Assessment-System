SELECT '=== tester ===' AS x;
SELECT t.id, t.name, t.paper_id, t.exam_id FROM el_tester t WHERE t.exam_id='1777345413187214616' AND t.name='高分';
SELECT '=== candidate ===' AS x;
SELECT c.id, c.name, c.paper_id, c.exam_id FROM el_candidate c WHERE c.exam_id='1777345413187214616' AND c.name='高分';
