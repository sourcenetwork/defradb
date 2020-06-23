package db

import (
	"fmt"
	"sync"

	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/merkle/crdt"

	"github.com/fxamacker/cbor/v2"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/jbenet/goprocess"
)

// Get a document from the given DocKey, return an error if we fail to retrieve
// the specified document.
// If the Key doesn't exist, return ErrDocumentNotFound
func (db *DB) Get(key key.DocKey) (*document.Document, error) {
	found, err := db.Exists(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrDocumentNotFound
	}

	return db.get(key)
}

// scans the database for the given document and all associated fields, returns document
func (db *DB) get(key key.DocKey) (*document.Document, error) {
	// To get the entire document, we dispatch a Query request to get all
	// keys with the prefix for the given DocKey.
	// This will return any and all keys under that prefix, which all fields
	// of the document exist, as well as any sub documents, etc
	q := query.Query{
		Prefix:   key.Key.String(),
		KeysOnly: false,
	}

	doc := document.NewWithKey(key)
	res, err := db.datastore.Query(q)
	defer res.Close()
	if err != nil {
		return nil, err
	}

	// dispatch collectors for each returned key/value pair.
	// Because our k/v layout utilizes multiple pairs to represent a given
	// field/value element of the document, and because the query isn't
	// guranteed to maintain any specific order, we need to asynchronisly
	// collect all the responses from the given channel, and dispatch them,
	// to the correct collector for the field they are apart of.
	// @todo: Investigate different field collector approach
	collector := newFieldCollector()
	for r := range res.Next() {
		// do we need to check r.Error here?
		if k := ds.NewKey(r.Key); k.Name() != "p" { // ignore priority key
			collector.dispatch(k.Type(), r.Entry)
		}
	}

	go func() {
		fmt.Println("-- Waiting for all queue processes to close --")
		collector.process.Close()
		fmt.Println("-- All process have completed and closed --")
	}()

	// waits for the collector to collate the necessary
	// k/v pairs, and returns a formatted field/value entry
	for {
		select {
		case fr := <-collector.results():
			fmt.Println("New field result:", fr)
			err = doc.SetAs(fr.name, fr.value, fr.ctype)
			if err != nil {
				return nil, err // wrap
			}
		case <-collector.process.Closed():
			fmt.Println("Collector process closed")
			return doc, nil
		}
	}
}

type fieldResult struct {
	// data [3][]byte // an array of size 3 of byte arrays to hold all the data we need per field pair
	// // The size is 1+number of values.
	// // 1 is from the field name
	// // and the remaining are all the values/metadata need for the field pair
	name  string
	value interface{}
	ctype crdt.Type
	err   error
}

// may want to abstract into an interface, so different implementations can decode the values as
// they need
type fieldCollector struct {
	queues         map[string]chan query.Entry
	fieldResultsCh chan fieldResult
	process        goprocess.Process
	sync.Mutex     // lock for queues map
}

func newFieldCollector() *fieldCollector {
	fc := fieldCollector{
		queues:         make(map[string]chan query.Entry),
		fieldResultsCh: make(chan fieldResult),
		process:        goprocess.Background(),
	}
	return &fc
}

func (c *fieldCollector) dispatch(field string, entry query.Entry) {
	c.Lock()
	q, ok := c.queues[field]
	if !ok {
		q = make(chan query.Entry)
		c.queues[field] = q
		fmt.Println("running new queue process")
		c.process.Go(func(p goprocess.Process) { // run queue inside its own process so we can control its exit condition
			c.runQueue(p, q)
		})
	}
	c.Unlock()
	q <- entry
}

// runs the loop for a given queue
// @todo: Handle subobject for fieldCollector
func (c *fieldCollector) runQueue(p goprocess.Process, q chan query.Entry) {
	collected := 0
	res := fieldResult{}
	for entry := range q {
		fmt.Println("Got a new entry on queue")
		k := ds.NewKey(entry.Key)
		// new entry, parse and insert
		if len(res.name) == 0 {
			res.name = k.Type()
			collected++
		}

		switch k.Name() {
		case "v": // main cbor encoded value
			err := cbor.Unmarshal(entry.Value, &res.value)
			if err != nil {
				res.err = err
				c.fieldResultsCh <- res
				close(q)
				p.Close()
			}
		case "ct": // cached crdt type, which is only a single byte, hence [0]
			res.ctype = crdt.Type(entry.Value[0])
		}

		// if weve completed all our tasks, close this queue/process down
		collected++
		fmt.Printf("Collected status: %d/3\n", collected)
		if collected == 3 { // maybe parameterize this constant
			fmt.Printf("Closing queue and process for %s...\n", res.name)
			c.fieldResultsCh <- res
			close(q)
			p.Close()
			fmt.Println("Closed queue and process for", res.name)
		}
	}
}

func (c *fieldCollector) results() <-chan fieldResult {
	return c.fieldResultsCh
}
