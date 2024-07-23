-- sql vaicājums testēšanai ar psql
-- psql <url> -f get-entry.sql -v entry_id=tezaurs_2024_2/vārds:1
WITH RECURSIVE senses2 AS (
	SELECT id, n, def, parent_id, entry_id
	FROM senses
	WHERE entry_id = :'entry_id'
	UNION
	SELECT s.id, s.n, s.def, s.parent_id, s.entry_id
	FROM senses s
	INNER JOIN senses2 s2 ON s.parent_id = s2.id
)
SELECT * FROM senses2;
