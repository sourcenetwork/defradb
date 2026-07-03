// smoke_kv exercises ExportDocKVs, ExportFieldMapping, ImportRawKVs, and
// ImportRawKVsWithMapping on an in-memory DefraDB node.
// Run from the repo root:  go run ./tests/manual/cmd/smoke_kv/
package main

import (
	"bytes"
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

	raw, _ := json.Marshal(map[string]any{"title": "Dune", "year": 1965, "author": "Herbert"})
	doc, err := client.NewDocFromJSON(ctx, raw, col.Version())
	if err != nil {
		fmt.Fprintln(os.Stderr, "NewDocFromJSON:", err)
		os.Exit(1)
	}
	if err := col.AddDocument(ctx, doc); err != nil {
		fmt.Fprintln(os.Stderr, "AddDocument:", err)
		os.Exit(1)
	}

	// ExportDocKVs — pass the docID explicitly.
	var buf bytes.Buffer
	n2, err := db.ExportDocKVs(ctx, "Book", []string{doc.ID().String()}, &buf, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ExportDocKVs:", err)
		os.Exit(1)
	}
	fmt.Printf("ExportDocKVs: %d pairs, %d bytes\n", n2, buf.Len()) //nolint:forbidigo

	// ExportFieldMapping
	mappingJSON, err := db.ExportFieldMapping(ctx, "Book")
	if err != nil {
		fmt.Fprintln(os.Stderr, "ExportFieldMapping:", err)
		os.Exit(1)
	}
	var mappingCheck map[string]any
	if err := json.Unmarshal(mappingJSON, &mappingCheck); err != nil {
		fmt.Fprintln(os.Stderr, "mapping JSON invalid:", err)
		os.Exit(1)
	}
	if _, ok := mappingCheck["collectionName"]; !ok {
		fmt.Fprintln(os.Stderr, "mapping missing collectionName key")
		os.Exit(1)
	}
	if _, ok := mappingCheck["fields"]; !ok {
		fmt.Fprintln(os.Stderr, "mapping missing fields key")
		os.Exit(1)
	}
	fmt.Printf("ExportFieldMapping: %s\n", string(mappingJSON)) //nolint:forbidigo

	// ImportRawKVs (idempotent re-import of same data)
	snap := buf.Bytes()
	n3, err := db.ImportRawKVs(ctx, bytes.NewReader(snap))
	if err != nil {
		fmt.Fprintln(os.Stderr, "ImportRawKVs:", err)
		os.Exit(1)
	}
	fmt.Printf("ImportRawKVs: %d pairs\n", n3) //nolint:forbidigo

	// ImportRawKVsWithMapping (identity remap — src/dst IDs are the same node)
	n4, err := db.ImportRawKVsWithMapping(ctx, bytes.NewReader(snap), mappingJSON)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ImportRawKVsWithMapping:", err)
		os.Exit(1)
	}
	fmt.Printf("ImportRawKVsWithMapping: %d pairs\n", n4) //nolint:forbidigo

	fmt.Println("kv: ok") //nolint:forbidigo
}
