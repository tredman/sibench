# sibench

## What

Simple benchmark for evaluating write performance of indexed collections in mongodb. The benchmark accepts the following test variables:

- numWorkers
- numDocs
- numIndexes
- numFields

And the following optional parameters:

- testHost (default: localhost)
- testDB (default: test_db)
- testColl (default: test_coll)

The program will create `numWorkers` workers to insert a total of `numDocs` documents into the specified collection/database. Each document will contain `numFields` fields with randomly generated int64 values. `numIndexes` of these fields will be indexed. 

Before running, the program drops `testDB` from the target server.

## Why

MongoDB supports a maximum of 64 indexes per collection. At Parse, it's common for some collections to have 20-30 indexes. We wanted some concrete data about the impact to write performance for each index added.

## How

```
$ go get github.com/tredman/sibench/cmd/sibench
```

To run:

```
NUM_WORKERS=10
NUM_INDEXES=1
NUM_DOCS=5000000
NUM_FIELDS=10
./sibench -num_indexes=${NUM_INDEXES} -num_fields ${NUM_FIELDS} -num_docs ${NUM_DOCS} -num_workers ${NUM_WORKERS}
```
