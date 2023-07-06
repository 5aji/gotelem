package db

import (
	"embed"
	"errors"
	"io"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
)

// embed the migrations into applications so they can update databases.

//go:embed migrations/*
var migrationsFs embed.FS

var migrationRegex = regexp.MustCompile(`^([0-9]+)_(.*)_(down|up)\.sql$`)

type Migration struct {
	Name     string
	Version  uint
	FileName string
}

type MigrationError struct {
}

// getMigrations returns a list of migrations, which are correctly index. zero is nil.
func getMigrations(files fs.FS) map[int]map[string]Migration {

	res := make(map[int]map[string]Migration) // version number -> direction -> migration.

	fs.WalkDir(files, ".", func(path string, d fs.DirEntry, err error) error {

		if d.IsDir() {
			return nil
		}
		m := migrationRegex.FindStringSubmatch(d.Name())
		if len(m) != 4 {
			panic("error parsing migration name")
		}
		migrationVer, _ := strconv.ParseInt(m[1], 10, 64)

		mig := Migration{
			Name:     m[2],
			Version:  uint(migrationVer),
			FileName: d.Name(),
		}

		var mMap map[string]Migration
		mMap, ok := res[int(migrationVer)]
		if !ok {
			mMap = make(map[string]Migration)
		}
		mMap[m[3]] = mig

		res[int(migrationVer)] = mMap

		return nil
	})
	return res
}

func RunMigrations(tdb *TelemDb) (finalVer int, err error) {

	currentVer, err := tdb.GetVersion()
	if err != nil {
		return
	}

	migrations := getMigrations(migrationsFs)

	// get a sorted list of versions.
	vers := make([]int, len(migrations))

	i := 0
	for k := range migrations {
		vers[i] = k
		i++
	}
	sort.Ints(vers)
	expectedVer := 1

	// check to make sure that there are no gaps (increasing by one each time)
	for _, v := range vers {
		if v != expectedVer {
			err = errors.New("missing update between")
			return
			// invalid
		}
		expectedVer = v + 1
	}

	finalVer = vers[len(vers)-1]
	// now apply the mappings based on current ver.

	tx, err := tdb.db.Begin()
	if err != nil {
		return
	}
	for v := currentVer + 1; v < finalVer; v++ {
		// attempt to get the "up" migration.
		mMap, ok := migrations[v]
		if !ok {
			err = errors.New("could not find migration for version")
			goto rollback
		}
		upMigration, ok := mMap["up"]
		if !ok {
			err = errors.New("could not get up migration")
			goto rollback
		}
		upFile, err := migrationsFs.Open(path.Join("migrations", upMigration.FileName))
		if err != nil {
			goto rollback
		}

		upStmt, err := io.ReadAll(upFile)
		if err != nil {
			goto rollback
		}
		// open the file name
		// execute the file.
		_, err = tx.Exec(string(upStmt))
		if err != nil {
			goto rollback
		}

	}
	// if all the versions applied correctly, update the PRAGMA user_version in the database.
	tx.Commit()
	err = tdb.SetVersion(finalVer)

	return
	// yeah, we use goto. Deal with it.
rollback:
	tx.Rollback()
	return
}
