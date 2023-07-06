CREATE TABLE "bus_events" (
	"ts"	INTEGER NOT NULL, -- timestamp, unix milliseconds
	"id"	INTEGER NOT NULL, -- can ID
	"name"	TEXT NOT NULL, -- name of base packet
	"data"	TEXT NOT NULL CHECK(json_valid(data)) -- JSON object describing the data, including index if any
);

CREATE INDEX "ids_timestamped" ON "bus_events" (
	"id",
	"ts"	DESC
);

CREATE INDEX "times" ON "bus_events" (
	"ts" DESC
);