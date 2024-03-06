ALTER TABLE "bus_events" ADD COLUMN idx GENERATED ALWAYS AS (json_extract(data, '$.idx')) VIRTUAL;
