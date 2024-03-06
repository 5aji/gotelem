-- this table shows when we started/stopped logging.
CREATE TABLE "drive_records" (
	"id"	INTEGER NOT NULL UNIQUE, -- unique ID of the drive.
	"start_time"	INTEGER NOT NULL, -- when the drive started
	"end_time"	INTEGER, -- when it ended, or NULL if it's ongoing.
	"note"	TEXT, -- optional description of the segment/experiment/drive
	PRIMARY KEY("id" AUTOINCREMENT),
	CONSTRAINT "duration_valid" CHECK(end_time is null or start_time < end_time)
);
