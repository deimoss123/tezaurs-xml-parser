-- sql vaicājums testēšanai ar psql
-- psql <url> -f get-entry.sql -v entry_id=tezaurs_2024_2/vārds:1
WITH RECURSIVE senses AS (
	SELECT id, n, def, parent_id, entry_id
	FROM sense
	WHERE entry_id = :'entry_id'
	UNION ALL
	SELECT s.id, s.n, s.def, s.parent_id, s.entry_id
	FROM sense s
	INNER JOIN senses ss ON s.parent_id = ss.id
)
SELECT * FROM senses;
