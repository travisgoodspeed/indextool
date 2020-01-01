
Howdy y'all,

IndexTool is a quick and dirty tool of mine for reviewing the indexing
of books written in LaTeX.  It isn't used to generate the index pages,
or to produce any typeset results at all; rather, it reviews the
manually indexed content to look for duplicates and mistakes.

The tool assumes that you have one file per chapter, that the same
word should not be indexed twice on one page or twice within a
chapter, and that each word is defined first in ASCII with any accent
marks or highlighting appearing later.

It first generates a SQLite3 database from the book's `.idx` file and
from its `.tex` source code, then successive calls are performed on
that database to allow for fast and interactive searches on individual
terms.  You can make manual queries on the database without trouble.

73 from Pizza Rat City,

--Travis Goodspeed


## Installation

IndexTool can be installed either with a traditional Unix `make clean
install` or by `go get github.com/travisgoodspeed/indextool`.

## Operation

IndexTool works by first generating a SQLite3 database of your LaTeX
source code (`*.tex`) and the output of mkindex (`*.idx`).  This needs
to be frequently regenerated as you correct your indexing, so you
ought to have a target in your `Makefile` that generates these markings.

```
index: book.idx *.tex
	indextool book.idx *.tex
```

Because indexing inherently involves the discretion of a human editor,
the tool's findings are all considered warnings, rather than errors.

After the database is generated, run the tool with no parameters to do
a quick sanity check of your indexing.  Anything reported at this
stage is likely a serious mistake, such as duplicate indexing.

```
x270% indextool
Duplicate entry 'PaX' on page 19.
Entry Capitalization: JavaScript or Javascript ?
x270%
```

Some queries are a bit more intensive, taking more than the tenth of a
second we'd like to budget for the default tests.  These are grouped
into "deep mode", which must be enabled separately using the `-d`
flag.

```
dell% indextool -d
...
Index Capitalization: BrainFuck or Brainfuck ?
Index Capitalization: CoreBoot or Coreboot ?
Index Capitalization: FastColl or Fastcoll ?
Index Capitalization: GNUPG or GnuPG ?
Index Capitalization: GameBoy or Gameboy ?
Index Capitalization: JavaScript or Javascript ?
Index Capitalization: Nintendo!GameBoy or Nintendo!Gameboy ?
Index Capitalization: PDF.JS or PDF.js ?
Index Capitalization: PostScript or Postscript ?
Index Capitalization: SHAttered or Shattered ?
Index Capitalization: WINE or Wine ?
Index Capitalization: X86 or x86 ?
dell% 
```

You can also perform specific queries.  For example, we can do a full
text search for PaX, to identify all files which contain the word but
have not indexed it.

```
x270% indextool -s PaX
Missing 'PaX' index in sample/ch2.tex.
x270% 
```

We can also list the entries--those that appear in the `.idx` file of
the book as it is actually rendered--by `indextool -l` or the
indices--those that appear anywhere in the source code code, even if
they aren't rendered--by `indextool -L`.  The distinction is handy in
that you might be update the `-L` listing without recompiling your
book; the `-l` listing is more accurate, and might be confined to just
the volume you are currently compiling.  For example, here is a listing
where the gameboy has been indexed multiple ways incorrectly.

```
dell% indextool -L | grep -i gameboy
GameBoy
GameBoy Advance
Gameboy
Nintendo!GameBoy
Nintendo!Gameboy
Super GameBoy
dell% 
```

## Raw Database Access

IndexTool's SQLite3 database is available with a default filename of
`indextool.db`.  You can open it to perform queries directly, if that
would be handy.  Search for `db.Query` in `indextool.go` for example
queries that might be handy.

The database is roughly like this.

```sql
/* From the .tex files.*/
CREATE TABLE indices(filename, name);

/* From the .idx files. */
CREATE TABLE entries(filename, name, page);

CREATE VIRTUAL TABLE tex using fts4(filename, body)
/* tex(filename,body) */;
CREATE TABLE IF NOT EXISTS 'tex_content'(docid INTEGER PRIMARY KEY, 'c0filename', 'c1body');
CREATE TABLE IF NOT EXISTS 'tex_segments'(blockid INTEGER PRIMARY KEY, block BLOB);
CREATE TABLE IF NOT EXISTS 'tex_segdir'(level INTEGER,idx INTEGER,start_block INTEGER,leaves_end_block INTEGER,end_block INTEGER,root BLOB,PRIMARY KEY(level, idx));
CREATE TABLE IF NOT EXISTS 'tex_docsize'(docid INTEGER PRIMARY KEY, size BLOB);
CREATE TABLE IF NOT EXISTS 'tex_stat'(id INTEGER PRIMARY KEY, value BLOB);
```


## Performance

Previously, this tool was significantly faster in apfs on MacOS than
in ext4 in Linux.  Rather than work around the performance issues
through ramdisks, the tool now uses `pragma synchronous = off;` to
disable all database safety.  This isn't considered a problem because
the database is regenerated with each indexing run.

