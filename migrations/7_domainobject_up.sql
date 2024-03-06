CREATE TABLE openmct_objects (
	data TEXT,
	key TEXT GENERATED ALWAYS AS (json_extract(data, '$.identifier.key')) VIRTUAL UNIQUE NOT NULL
);
-- fast key-lookup
CREATE INDEX openmct_key on openmct_objects(key);
