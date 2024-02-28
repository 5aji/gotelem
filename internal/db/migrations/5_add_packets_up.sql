CREATE TABLE "packet_definitions" (
	"name" TEXT NOT NULL,
	"description" TEXT,
	"id" INTEGER NOT NULL
);

CREATE TABLE "field_definitions" (
	"name" TEXT NOT NULL,
	"subname" TEXT, -- if the data type is a bitfield, we can use subname to identify the bit.
	"packet_name" TEXT NOT NULL,
	"type" TEXT NOT NULL,
	FOREIGN KEY("packet_name") REFERENCES packet_definitions(name)
);
