-- Purge lists which do not have any participant 
DELETE FROM lists WHERE id IN (SELECT id FROM lists l WHERE NOT EXISTS(SELECT * FROM lists_users_pivot p WHERE p.list_id = l.id));
