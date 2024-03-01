CREATE TABLE "bus_events" (
	"ts"	INTEGER NOT NULL, -- timestamp, unix milliseconds
	"name"	TEXT NOT NULL, -- name of base packet
	"data"	JSON NOT NULL CHECK(json_valid(data)) -- JSON object describing the data, including index if any
);

CREATE INDEX "ids_timestamped" ON "bus_events" (
	"name",
	"ts"	DESC
);

CREATE INDEX "times" ON "bus_events" (
	"ts" DESC
);
