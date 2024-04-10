SET pg_trgm.similarity_threshold = 0.2;

SELECT id, def  
FROM senses
WHERE def %> :'query'
ORDER BY def <-> :'query' ASC
LIMIT 100;

-- SELECT id, ts_rank(def_tsvector, query) AS rank, ts_headline('english', def, query) AS headline
-- FROM senses, plainto_tsquery('english', :'query') query
-- WHERE def_tsvector @@ query
-- ORDER BY rank DESC
-- LIMIT 5;

-- SELECT id, ts_rank(def_tsvector, query) + similarity(def, 'your search term') AS rank, ts_headline('english', def, query) AS headline
-- FROM senses, plainto_tsquery('english', :'query') query
-- WHERE def_tsvector @@ query OR def % :'query'
-- ORDER BY rank DESC
-- LIMIT 10;
