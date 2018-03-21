


all: indextool

indextool: *.go
	go build
fmt:
	go fmt

clean:
	rm -f indextool indextool.db *~ */*~

run: test
test: all indextool.db
	./indextool -d   #Deep scan.

indextool.db: all
	./indextool sample/*.idx sample/*.tex   #Generate the database.

#Installs prereq packages.
get:
	go get github.com/mattn/go-sqlite3
	go get gopkg.in/cheggaaa/pb.v2


