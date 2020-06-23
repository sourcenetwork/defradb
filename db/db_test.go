package db

import (
	"testing"

	"github.com/sourcenetwork/defradb/document/key"

	"github.com/sourcenetwork/defradb/document"

	"github.com/stretchr/testify/assert"
)

func newMemoryDB() (*DB, error) {
	opts := &Options{
		Store: "memory",
		Memory: MemoryOptions{
			Size: 1024 * 1000,
		},
	}

	return NewDB(opts)
}

func TestNewDB(t *testing.T) {
	opts := &Options{
		Store: "memory",
		Memory: MemoryOptions{
			Size: 1024 * 1000,
		},
	}

	_, err := NewDB(opts)
	if err != nil {
		t.Error(err)
	}
}

func TestDBSaveSimpleDocument(t *testing.T) {
	db, _ := newMemoryDB()

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = db.Save(doc)
	if err != nil {
		t.Error(err)
	}

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	weight, err := doc.Get("Weight")
	assert.NoError(t, err)

	assert.Equal(t, "John", name)
	assert.Equal(t, int64(21), age)
	assert.Equal(t, 154.1, weight)

	_, err = doc.Get("DoesntExist")
	assert.Error(t, err)

	// db.printDebugDB()
}

func TestDBUpdateDocument(t *testing.T) {
	db, _ := newMemoryDB()

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = db.Save(doc)
	if err != nil {
		t.Error(err)
	}

	// update fields
	doc.Set("Name", "Pete")
	doc.Delete("Weight")

	weightField := doc.Fields()["Weight"]
	weightVal, _ := doc.GetValueWithField(weightField)
	assert.True(t, weightVal.IsDelete())

	err = db.Update(doc)

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	weight, err := doc.Get("Weight")
	assert.NoError(t, err)

	assert.Equal(t, "Pete", name)
	assert.Equal(t, int64(21), age)
	assert.Nil(t, weight)

	// fmt.Println("\n--")
	// db.printDebugDB()
}

func TestDBUpdateNonExistingDocument(t *testing.T) {
	db, _ := newMemoryDB()

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	if err != nil {
		t.Error(err)
		return
	}

	err = db.Update(doc)
	assert.Error(t, err)
}

func TestDBUpdateExistingDocument(t *testing.T) {
	db, _ := newMemoryDB()

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = db.Save(doc)
	assert.NoError(t, err)

	testJSONObj = []byte(`{
		"_key": "bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d",
		"Name": "Pete",
		"Age": 31
	}`)

	doc, err = document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = db.Update(doc)
	assert.NoError(t, err)

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	// weight, err := doc.Get("Weight")
	// assert.NoError(t, err)

	assert.Equal(t, "Pete", name)
	assert.Equal(t, int64(31), age)
}

func TestDBGetDocument(t *testing.T) {
	db, _ := newMemoryDB()

	testJSONObj := []byte(`{
		"Name": "John",
		"Age": 21,
		"Weight": 154.1
	}`)

	doc, err := document.NewFromJSON(testJSONObj)
	assert.NoError(t, err)

	err = db.Save(doc)
	assert.NoError(t, err)

	key, err := key.NewFromString("bae-09cd7539-9b86-5661-90f6-14fbf6c1a14d")
	assert.NoError(t, err)
	doc, err = db.Get(key)
	assert.NoError(t, err)

	// value check
	name, err := doc.Get("Name")
	assert.NoError(t, err)
	age, err := doc.Get("Age")
	assert.NoError(t, err)
	weight, err := doc.Get("Weight")
	assert.NoError(t, err)

	assert.Equal(t, "Pete", name)
	assert.Equal(t, int64(31), age)
	assert.Equal(t, 154.1, weight)
}
