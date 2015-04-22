package main

import (
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	numWorkers int
	numFields  int
	numIndexes int
	numDocs    int
	testHost   string
	testDB     string
	testColl   string
	fields     []string
)

func init() {
	flag.IntVar(&numWorkers,
		"num_workers", 10,
		"Number of worker threads. num_docs will be divided by this "+
			"number, and each worker will execute that many threads.")
	flag.IntVar(&numFields,
		"num_fields", 10,
		"Number of fields to create. Only num_indexes fields will be indexed. "+
			"Each field will receive the same data (a randomized string) but the "+
			"data will change for each doc.")
	flag.IntVar(&numIndexes,
		"num_indexes", 0,
		"Number of fields to index. Should be < num_fields but mongo won't "+
			"complain if it's higher - it will just index non-existent fields.")
	flag.IntVar(&numDocs, "num_docs", 1000, "Number of total docs to insert.")
	flag.StringVar(&testHost, "test_host", "localhost",
		"Hostname of mongo instance to connect to. Standalone is assumed.")
	flag.StringVar(&testDB, "test_db", "test_db",
		"Name of database to insert docs into.")
	flag.StringVar(&testColl, "test_coll", "test_coll",
		"Name of collection to insert docs into.")

	flag.Parse()

	fields = make([]string, numFields, numFields)
	for i := 0; i < numFields; i++ {
		fields[i] = fmt.Sprintf("field_%d", i)
	}

}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// log args at start of run
	fmt.Printf("Host: %s\n", testHost)
	fmt.Printf("DB: %s\n", testDB)
	fmt.Printf("Collection: %s\n", testColl)
	fmt.Printf("Workers: %d\n", numWorkers)
	fmt.Printf("Fields: %d\n", numFields)
	fmt.Printf("Indexes: %d\n", numIndexes)

	session, err := mgo.Dial(testHost)
	if err != nil {
		fmt.Println(err)
		return
	}
	db := session.DB(testDB)
        // throw away and recreate the database/collection each time
	db.DropDatabase()
	c := db.C(testColl)
	err = c.Create(&mgo.CollectionInfo{})
	if err != nil {
		fmt.Println(err)
		return
	}
        // indexes are created up front
	for i := 0; i < numIndexes; i++ {
		c.EnsureIndex(mgo.Index{Key: []string{fmt.Sprintf("field_%d", i)}})
	}

	wg := sync.WaitGroup{}
	var opCounter int64
	var errCounter int64

	stop := make(chan bool)
	defer func() {
		stop <- true
	}()
	// print stats periodically
	go func() {
		opsLast := int64(0)
		errsLast := int64(0)

		for {
			fmt.Printf("total ops: %d, errors: %d, ops/sec: %d, errors/sec: %d\n",
				opCounter, errCounter, opCounter-opsLast, errCounter-errsLast)
			opsLast = opCounter
			errsLast = errCounter

			select {
			case <-stop:
				return
			default:
				time.Sleep(time.Second)
			}
		}
	}()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()
			session, err := mgo.Dial(testHost)
			if err != nil {
				panic(fmt.Sprintf("worker %d failed to connect, bailing."))
			}
			db := session.DB(testDB)
			c := db.C(testColl)

			for i := 0; i < (numDocs / numWorkers); i++ {
				data := rand.Int63()
				doc := bson.M{}
				for i := 0; i < numFields; i++ {
					doc[fields[i]] = data
				}
				err = c.Insert(doc)
				if err != nil {
					fmt.Println(err)
					atomic.AddInt64(&errCounter, 1)
				}
				atomic.AddInt64(&opCounter, 1)
			}
		}(i)
	}

	wg.Wait()
}
