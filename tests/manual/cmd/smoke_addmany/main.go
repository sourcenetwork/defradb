// smoke_addmany exercises SaveManyDocuments on an in-memory DefraDB node.
// Run from the repo root:  go run ./tests/manual/cmd/smoke_addmany/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
)

func main() {
	ctx := context.Background()

	opts := options.Node().
		SetDisableP2P(true).
		SetDisableAPI(true).
		SetEnableDevelopment(true).
		Store().SetBadgerInMemory(true).
		Node()
	n, err := node.New(ctx, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "node.New:", err)
		os.Exit(1)
	}
	if err := n.Start(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "n.Start:", err)
		os.Exit(1)
	}
	defer n.DB.Close()
	db := n.DB

	_, err = db.AddCollection(ctx, `type Book { title: String, year: Int, author: String }`)
	if err != nil {
		fmt.Fprintln(os.Stderr, "AddCollection:", err)
		os.Exit(1)
	}

	col, err := db.GetCollectionByName(ctx, "Book")
	if err != nil {
		fmt.Fprintln(os.Stderr, "GetCollectionByName:", err)
		os.Exit(1)
	}

	books := []map[string]any{
		{"title": "Dune", "year": 1965, "author": "Herbert"},
		{"title": "Neuromancer", "year": 1984, "author": "Gibson"},
		{"title": "Snow Crash", "year": 1992, "author": "Stephenson"},
		{"title": "Foundation", "year": 1951, "author": "Asimov"},
		{"title": "1984", "year": 1949, "author": "Orwell"},
	}
	docs := make([]*client.Document, len(books))
	for i, b := range books {
		raw, _ := json.Marshal(b)
		doc, err := client.NewDocFromJSON(ctx, raw, col.Version())
		if err != nil {
			fmt.Fprintln(os.Stderr, "NewDocFromJSON:", err)
			os.Exit(1)
		}
		docs[i] = doc
	}

	if err := col.SaveManyDocuments(ctx, docs); err != nil {
		fmt.Fprintln(os.Stderr, "SaveManyDocuments:", err)
		os.Exit(1)
	}

	res := db.ExecRequest(ctx, `{Book{_docID,title}}`)
	if len(res.GQL.Errors) > 0 {
		fmt.Fprintln(os.Stderr, "ExecRequest:", res.GQL.Errors)
		os.Exit(1)
	}
	// Normalize via JSON round-trip: GQL data uses []map[string]any, not []any.
	b, _ := json.Marshal(res.GQL.Data)
	var result map[string][]map[string]any
	if err := json.Unmarshal(b, &result); err != nil {
		fmt.Fprintln(os.Stderr, "result unmarshal:", err)
		os.Exit(1)
	}
	readable := len(result["Book"])
	fmt.Printf("SaveManyDocuments: saved=%d readable=%d\n", len(docs), readable)
	if readable != len(docs) {
		fmt.Fprintf(os.Stderr, "count mismatch: expected %d, got %d\n", len(docs), readable)
		os.Exit(1)
	}
	fmt.Println("SaveManyDocuments: ok")
}
