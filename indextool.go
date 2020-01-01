// This is a handy little tool for auditing indexes of large LaTeX
// documents.  See the README for examples, and please send me one
// copy of your book if you find it to be useful.

// Inputs are parsed by regular expression, so they must have either
// each \index or \indexentry on a single line.  The tool is confused
// by access marks, but you ought to have the ASCII form of the index
// first regardless.

// --Travis

package main

import (
	"bufio"
	"flag"
	"fmt"
	//"os"
	//"io"
	"io/ioutil"
	//"strconv"
	"regexp"
	"strconv"
	"strings"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"gopkg.in/cheggaaa/pb.v2"
)

//Set to true in verbose mode.
var verbose bool

//Set true to scan for all missing entries.
var deepmode bool

//Fails on an error, but prints it first.
func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Parse a whole file, rather inefficiently.
func parseFile(filename string) {
	if verbose {
		fmt.Printf("# Parsing %s\n",
			filename)
	}

	//This regex matches indexes in LaTeX.
	indexre, err := regexp.Compile("\\\\index{(.*)}")
	check(err)

	//This regex matches indexentries in .idx
	entryre, err := regexp.Compile("\\\\indexentry{(.*)}{(.*)}")
	check(err)

	//First we slurp the whole file into RAM.  This is a bit wasteful.
	dat, err := ioutil.ReadFile(filename)
	check(err)

	//Convert the bytes to a string, then insert that for full text search.
	s := string(dat)
	insertTex(filename, s)

	//Create a scanner around the input, converted to a string.
	scanner := bufio.NewScanner(strings.NewReader(s))

	// Split into lines by default.
	scanner.Split(bufio.ScanLines)
	// Validate the input
	for scanner.Scan() {
		line := scanner.Text()

		index := indexre.FindStringSubmatch(line)
		if len(index) == 2 {
			insertIndex(filename, index[1])
		}

		entry := entryre.FindStringSubmatch(line)
		if len(entry) == 3 {
			page, _ := strconv.Atoi(entry[2])
			insertEntry(filename, entry[1], page)
		}
	}
}

//SQLite3 database connection.
var db *sql.DB

//Open a database, then creates the tables if they don't already eixst.
func opendb(filename string) {
	database, err := sql.Open("sqlite3", filename)
	check(err)
	db = database

	initdb()
}

//Drop the old tables, then creates new ones.
func dropdb() {
	db.Exec("drop table if exists indices;")
	db.Exec("drop table if exists entries;")
	db.Exec("drop table if exists tex;")

	initdb()
}

//Create the database tables.
func initdb() {
	//Simple tables for the indices (.idx) and entries (.tex)
	db.Exec("pragma synchronous = off;")
	db.Exec("create table if not exists indices(filename, name);")
	db.Exec("create table if not exists entries(filename, name, page);")

	//Full text search tables, for recognizing missing index entries.
	db.Exec("create virtual table if not exists  tex using fts4(filename, body);")

}

//Inserts a record for an entire file.
func insertTex(filename string, body string) {
	//We don't include the .idx files in the full text searches.
	if strings.HasSuffix(filename, ".idx") {
		return
	}

	if verbose {
		fmt.Printf("# Full Text Search of '%s'.\n", filename)
	}
	db.Exec("insert into tex (filename, body) values (?,?);",
		filename, body)
}

//Inserts a record for a \index{} line.
func insertIndex(filename string, name string) {
	if verbose {
		fmt.Printf("# Index to '%s' in %s.\n", name, filename)
	}
	db.Exec("insert into indices (filename, name) values (?,?);",
		filename, name)
}

//Inserts a record for an \\indexentry{} line.
func insertEntry(filename string, name string, page int) {
	if verbose {
		fmt.Printf("# IndexEntry to '%s' at page %d.\n", name, page)
	}
	db.Exec("insert into entries (filename, name, page) values (?,?,?);",
		filename, name, page)
}

//Prints missing indexes of a given string.
func printMissing(word string) {
	if verbose {
		fmt.Printf("# Searching for missing entries to %s.\n",
			word)
	}

	rows, err := db.Query("select filename from tex where body match ? and filename not in (select filename from indices where name like '%'||?||'%');",
		word, word)
	check(err)

	defer rows.Close()
	for rows.Next() {
		var filename string
		err = rows.Scan(&filename)
		check(err)

		fmt.Printf("Missing '%s' index in %s.\n",
			word, filename)
	}
	check(rows.Err())
}

//Prints duplicate entries on the same page.
func printEntryDuplicates() {
	rows, err := db.Query("select filename, name, page, count(*) from entries group by filename, name, page having count(*)>1;")
	check(err)

	defer rows.Close()
	for rows.Next() {
		var filename string
		var name string
		var page int
		var count int
		err = rows.Scan(&filename, &name, &page, &count)
		check(err)

		fmt.Printf("Duplicate entry '%s' on page %d.\n",
			name, page)
	}
	check(rows.Err())
}

//Prints all entries.
func printEntryList() {
	rows, err := db.Query("select distinct name from entries order by name asc;")
	check(err)

	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		check(err)

		fmt.Printf("%s\n", name)
	}
	check(rows.Err())
}

//Prints all Indicess.
func printIndexList() {
	rows, err := db.Query("select distinct name from indices order by name asc;")
	check(err)

	defer rows.Close()
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		check(err)

		fmt.Printf("%s\n", name)
	}
	check(rows.Err())
}

//Prints duplicate indices in the same .tex file.
func printIndexDuplicates() {
	rows, err := db.Query("select filename, name, count(*) from indices group by filename, name having count(*)>1;")
	check(err)

	defer rows.Close()
	for rows.Next() {
		var filename string
		var name string
		var count int
		err = rows.Scan(&filename, &name, &count)
		check(err)

		fmt.Printf("Duplicate entry '%s' in %s.\n",
			name, filename)
	}
	check(rows.Err())
}

//Main entry point.
func main() {
	//Parameters.  Only add one here as it is implemented in code.
	//mindistPtr := flag.Int("mindist", 5, "Min distance of two index entries.")
	//dupcharPtr := flag.Int("dupdist", 3, "Character count before duplicate distance.")
	verbosePtr := flag.Bool("v", false, "Verbose mode.")
	deepmodePtr := flag.Bool("d", false, "Deep scan mode.")
	entrylistmodePtr := flag.Bool("l", false, "Entry list mode.")
	indexlistmodePtr := flag.Bool("L", false, "Index list mode.")
	databasenamePtr := flag.String("f", "indextool.db", "Database filename.")
	searchPtr := flag.String("s", "", "Search for missing word entries.")
	flag.Parse()

	//Record some globals
	verbose = *verbosePtr
	deepmode = *deepmodePtr

	//Re-initialize the database when given files.
	if len(flag.Args()) > 0 {
		//Initialize a fresh database.
		opendb(*databasenamePtr)
		dropdb() //Flush the old tables.

		//Handle the input files.
		icount := len(flag.Args())
		bar := pb.StartNew(icount)
		for i := 0; i < icount; i++ {
			bar.Increment()
			parseFile(flag.Args()[i])
		}
		bar.Finish()

	} else { //Queries are only accepted without files.
		opendb(*databasenamePtr)

		//Print the results
		printEntryDuplicates() //Revealed entries on the same page.
		printIndexDuplicates() //Identical entries from the same file.

		//Deep mode uses a full-text search to provide a list of
		//entries (.idx lines) which are used but not indexed in a
		//given file.  There will be some false positives, of course,
		//as some terms are casually references and oughtn't be in the
		//index.  It might take a while.
		if deepmode {
			fmt.Printf("Deep mode ain't quite working yet.\n")

			//select distinct name from entries;
		}

		// -l, List mode dumps a list of all entries, for
		// quickly finding duplicates or near-duplicates
		// visually.
		if *entrylistmodePtr {
			//select distinct name from entries order by name asc;
			printEntryList()
		}

		// -L, longer list than above.
		if *indexlistmodePtr {
			//select distinct name from indices order by name asc;
			printIndexList()
		}

		//Search for a specific missing entry.
		if len(*searchPtr) > 0 {
			printMissing(*searchPtr)
		}

		//Done.
		db.Close()
	}
}
