
Howdy y'all,

IndexTool is a quick and dirty tool of mine for reviewing the indexing
of books written in LaTeX.  It isn't used to generate the index pages,
or to produce any typeset results at all; rather, it reviews the
manually indexed content to look for duplicates and mistakes.

The tool assumes that you have one file per chapter, that the same
word should not be indexes twice on one page, and that each word is
defined first in ASCII with any accent marks or highlighting appearing
later.  It first generates a database, then successive calls are
performed on that database to allow for fast and interactive searches
on individual terms.

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
x270%
```

You can also perform specific queries.  For example, we can do a full
text search for PaX, to identify all files which contain the word but
have not indexed it.

```
x270% indextool -s PaX
Duplicate entry 'PaX' on page 19.
Missing 'PaX' index in sample/ch2.tex.
x270% 
```

## Raw Database Access

IndexTool's SQLite3 database is available with a default filename of
`indextool.db`.  You can open it to perform queries directly, if that
would be handy.  Search for `db.Query` in `indextool.go` for example
queries that might be handy.


## Performance

On Linux, very large datasets take a long time on the `ext4`
filesystem, even on a modern SSD.  Considerably better performance can
be had by storing `indextool.db` on a `tmpfs` partition.

On Mac OS X, `apfs` is thirty times faster than `ext4`.  On this
platform, I don't bother with ramdisks.

