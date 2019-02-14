package tables

import (
	"sync"
	"github.com/hewiefreeman/GopherGameDB/helpers"
)

var (
	tablesMux      sync.Mutex
	logPersistTime int = 30
	tables         map[string]*table = make(map[string]*table)
)

type tableSchema struct {
	items map[string]tableSchemaItem
}

type tableSchemaItem struct {
	index       int
	unique      bool
	grouped     bool
	groupedEntries []interface{}
}

type table struct {
	logFile       string
	persistFolder string
	partitionMax  int
	maxEntries    int
	schema        tableSchema

	eMux    sync.Mutex
	entries map[string]*tableEntry

	iMux  sync.Mutex
	index []*tableEntry

	uMux       sync.Mutex
	uniqueEntries map[int]*uniqueTableEntry

	gMux        sync.Mutex
	groupedEntries map[int]*groupedTableEntry

	pMux   sync.Mutex
	fileOn int
	lineOn int
}

type tableEntry struct {
	key string

	persistFile  int
	persistIndex int

	mux   sync.Mutex
	entry []interface{}
}

// Database statuses
const (
	statusSettingUp = iota
	statusHealthy
	statusReplicationFailure
	statusOffline
)

// File/folder prefixes
const (
	prefixTableLogFile           = "log-"
	prefixTableDataFolder        = "data-"
	prefixTableDataPartitionFile = "part-"
)

// Defaults
const (
	defaultPartitionMax = 1500
	defaultConfig = "{\"dbName\":\"db\",\"replica\":false,\"readOnly\":false,\"logPersistTime\":30,\"replicas\":[],\"balancers\":[],\"tables\":[]}"
)

/////////////////////////////////////////////////////////////////////////////////////////////////
//   tableSchema   //////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func newTableSchema(items []string, unique []int, grouped map[int][]interface{}) (tableSchema, int) {
	if len(items) == 0 {
		return tableSchema{}, helpers.ErrorSchemaItemsRequired
	}
	var s tableSchema
	s.items = make(map[string]tableSchemaItem)
	// Go through items
	for i := 0; i < len(items); i++ {
		se := tableSchemaItem{index: i}
		// check for unique
		for j := 0; j < len(unique); j++ {
			if unique[j] == i {
				se.unique = true
			}
		}
		// check for grouped
		if !se.unique {
			if v, ok := grouped[i]; ok {
				// check if any of the values aren't hashable
				for j := 0; j < len(v); j++ {
					if !helpers.IsHashable(v[j]) {
						return tableSchema{}, helpers.ErrorUnhashableGroupValue
					}
				}
				se.grouped = true
				se.groupedEntries = v
			}
		}
		s.items[items[i]] = se
	}
	return s, 0
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   table   ////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func createTable(name string, maxEntries int, schema tableSchema, fileOn int, lineOn int, partitionMax int) int {
	if len(name) == 0 {
		return helpers.ErrorTableNameRequired
	} else if tableExists(name) {
		return helpers.ErrorTableExists
	}

	//default partitionMax
	if partitionMax <= 0 {
		partitionMax = defaultPartitionMax
	}

	//cant be under 0
	if fileOn < 0 {
		fileOn = 0
	}
	if lineOn < 0 {
		lineOn = 0
	}

	t := table{	logFile: prefixTableLogFile+name,
				persistFolder: prefixTableDataFolder+name,
				partitionMax: partitionMax,
				maxEntries: maxEntries,
				schema: schema,
				entries: make(map[string]*tableEntry),
				index: []*tableEntry{},
				uniqueEntries: make(map[int]*uniqueTableEntry),
				groupedEntries: make(map[int]*groupedTableEntry),
				fileOn: fileOn,
				lineOn: lineOn }

	// Apply schema
	for _, v := range schema.items {
		if v.unique {
			ute := uniqueTableEntry{vals: make(map[interface{}]*tableEntry)}
			t.uniqueEntries[v.index] = &ute
		} else if v.grouped {
			gte := groupedTableEntry{groups: make(map[interface{}]*tableEntryGroup)}
			for i := 0; i < len(v.groupedEntries); i++ {
				teg := tableEntryGroup{entries: []*tableEntry{}}
				gte.groups[v.groupedEntries[i]] = &teg
			}
			t.groupedEntries[v.index] = &gte
		}
	}

	//
	tablesMux.Lock()
	tables[name] = &t
	tablesMux.Unlock()

	// !!!!!! make new folder on system for persisting data & create log file & update config file

	return 0
}

func deleteTable(name string) int {
	if len(name) == 0 {
		return helpers.ErrorTableNameRequired
	} else if !tableExists(name) {
		return helpers.ErrorTableDoesntExist
	}

	tablesMux.Lock()
	delete (tables, name)
	tablesMux.Unlock()

	// !!!!!! delete data folder from system & delete log file & update config file

	return 0
}

func getTable(n string) *table {
	tablesMux.Lock()
	t := tables[n]
	tablesMux.Unlock()
	return t
}

func tableExists(n string) bool {
	tablesMux.Lock()
	var t bool = (tables[n] != nil)
	tablesMux.Unlock()
	return t
}

func (t *table) size() int {
	(*t).eMux.Lock()
	s := len((*t).entries)
	(*t).eMux.Unlock()
	return s
}

func (t *table) getEntry(n string) *tableEntry {
	(*t).eMux.Lock()
	e := (*t).entries[n]
	(*t).eMux.Unlock()
	return e
}

func (t *table) getUniqueEntry(index int) *uniqueTableEntry {
	t.uMux.Lock()
	u := t.uniqueEntries[index]
	t.uMux.Unlock()
	return u
}

func (t *table) getGroupedEntry(index int) *groupedTableEntry {
	t.gMux.Lock()
	g := t.groupedEntries[index]
	t.gMux.Unlock()
	return g
}

/////////////////////////////////////////////////////////////////////////////////////////////////
//   Misc methods   /////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

func makeEntryMap(entry []interface{}, schema tableSchema, sel []string) map[string]interface{} {
	entryMap := make(map[string]interface{})
	selLen := len(sel)
	for k, v := range schema.items {
		if selLen > 0 {
			for i := 0; i < selLen; i++ {
				if sel[i] == k {
					entryMap[k] = entry[v.index]
				}
			}
		} else {
			entryMap[k] = entry[v.index]
		}
	}
	return entryMap
}
