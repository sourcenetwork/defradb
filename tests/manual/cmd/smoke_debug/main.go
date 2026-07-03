package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
)

func main() {
	ctx := context.Background()
	opts := options.Node().
		SetDisableP2P(true).SetDisableAPI(true).SetEnableDevelopment(true).
		Store().SetBadgerInMemory(true).Node()
	n, err := node.New(ctx, opts)
	if err != nil {
		panic(err)
	}
	if err := n.Start(ctx); err != nil {
		panic(err)
	}
	defer n.DB.Close()
	db := n.DB

	_, err = db.AddCollection(ctx, `type Book { title: String, year: Int, author: String }`)
	fmt.Println("AddCollection err:", err) //nolint:forbidigo

	col, err := db.GetCollectionByName(ctx, "Book")
	fmt.Println("GetCollectionByName err:", err) //nolint:forbidigo

	// Same 5 books as smoke_addmany
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
		doc, docErr := client.NewDocFromJSON(ctx, raw, col.Version())
		if docErr != nil {
			panic("NewDocFromJSON: " + docErr.Error())
		}
		docs[i] = doc
	}

	err = col.SaveManyDocuments(ctx, docs)
	fmt.Println("SaveManyDocuments err:", err) //nolint:forbidigo

	res := db.ExecRequest(ctx, `{Book{_docID,title}}`)
	fmt.Println("ExecRequest errors:", res.GQL.Errors) //nolint:forbidigo
	fmt.Printf("ExecRequest data type: %T\n", res.GQL.Data) //nolint:forbidigo
	fmt.Printf("ExecRequest data: %v\n", res.GQL.Data) //nolint:forbidigo

	m, ok := res.GQL.Data.(map[string]any)
	fmt.Printf("type assert ok=%v\n", ok) //nolint:forbidigo
	if ok {
		for k, v := range m {
			fmt.Printf("  key=%q type=%T\n", k, v) //nolint:forbidigo
		}
		rows, _ := m["Book"].([]any)
		fmt.Printf("  rows len=%d\n", len(rows)) //nolint:forbidigo
	}
}
