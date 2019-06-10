package schema

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"strconv"
	"strings"
	"time"
	"sync"
)

// Filter for queries
type Filter struct {
	item interface{}
	methods []string
	dbEntryData *interface{}
	schemaItems []*SchemaItem
	schemaItemOn int
	uMux *sync.Mutex
	uniqueVals *map[string]map[string]bool
}

// Method names
const (
	MethodOperatorAdd = "*add"
	MethodOperatorSub = "*sub"
	MethodOperatorMul = "*mul"
	MethodOperatorDiv = "*div"
	MethodOperatorMod = "*mod"
	MethodAppend      = "*append"
	MethodAppendAt    = "*append["
	MethodAppendAtFin = "]"
	MethodPrepend     = "*prepend"
	MethodDelete      = "*delete"
)

// Item type query filters - Initialized when the first Schema is made (see New())
var (
	queryFilters map[string]func(*Filter)(int)
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   QUERY FILTER   /////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewFilter(item interface{}, methods []string, dbEntryData *interface{}, schemaItems []*SchemaItem, uMux *sync.Mutex, uniqueVals *map[string]map[string]bool) Filter {
	return Filter{
		item: item,
		methods: methods,
		dbEntryData: dbEntryData,
		schemaItems: schemaItems,
		uMux: uMux,
		uniqueVals: uniqueVals,
	}
}

// QueryItemFilter takes in an item from a query, and filters/checks it for format/completion against the cooresponding SchemaItem data type.
func QueryItemFilter(filter *Filter) int {
	if filter.item == nil {
		if len(filter.methods) > 0 {
			return helpers.ErrorInvalidMethodParameters
		}
		// Get default value
		defaultVal, defaultErr := defaultVal(filter.schemaItems[filter.schemaItemOn])
		if defaultErr != 0 {
			return defaultErr
		}
		if filter.schemaItemOn == 0 {
			(*(*filter).dbEntryData) = defaultVal
		} else {
			filter.item = defaultVal
			filter.schemaItemOn--
		}
		return 0
	} else {
		// Run type filter
		iTypeErr := queryFilters[filter.schemaItems[filter.schemaItemOn].typeName](filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		return 0
	}
}

func Format(tableItem interface{}, itemType *SchemaItem) interface{} {
	if itemType.typeName == ItemTypeTime {
		// Format Time types
		t := tableItem.(time.Time)
		it := itemType.iType.(TimeItem)
		return t.Format(it.format)
	}
	return tableItem
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ITEM TYPE METHODS   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func GetQueryItemMethods(itemName string) (string, []string) {
	var itemMethods []string
	if strings.Contains(itemName, ".") {
		ml := strings.Split(itemName, ".")
		itemMethods = ml[1:]
		itemName = ml[0]
	}
	return itemName, itemMethods
}

func applyNumberMethods(numbs []interface{}, methods []string, dbEntryData float64) (float64, int) {
	// Must have same amount of numbers in array as methods to use on them
	if len(numbs) != len(methods) {
		return 0, helpers.ErrorInvalidMethodParameters
	}
	for i, numb := range numbs {
		// Check numb type
		if cNumb, ok := makeFloat(numb); ok {
			op := methods[i]
			switch op {
			case MethodOperatorAdd:
				dbEntryData = dbEntryData + cNumb

			case MethodOperatorSub:
				dbEntryData = dbEntryData - cNumb

			case MethodOperatorMul:
				dbEntryData = dbEntryData * cNumb

			case MethodOperatorDiv:
				dbEntryData = dbEntryData / cNumb

			case MethodOperatorMod:
				dbEntryData = float64(int(dbEntryData) % int(cNumb))

			default:
				return 0, helpers.ErrorInvalidMethod
			}
		} else {
			return 0, helpers.ErrorInvalidMethodParameters
		}
	}
	return dbEntryData, 0
}

func applyStringMethods(strs []interface{}, methods []string, dbEntryData string) (string, int) {
	// Must have same amount of strings in array as methods to use on them
	if len(strs) != len(methods) {
		return "", helpers.ErrorInvalidMethodParameters
	}
	for i, str := range strs {
		// Check string type
		if cStr, ok := str.(string); ok {
			op := methods[i]
			switch op {
			case MethodOperatorAdd, MethodAppend:
				dbEntryData = dbEntryData + cStr
				continue

			case MethodPrepend:
				dbEntryData = cStr + dbEntryData
				continue
			}
			// Check for append at index method
			if len(methods[i]) >= 10 && methods[i][:8] == MethodAppendAt && methods[i][len(methods[i])-1:len(methods[i])] == MethodAppendAtFin {
				// Convert the text inside brackets to int
				j, jErr := strconv.Atoi(methods[i][8 : len(methods[i])-1])
				if jErr != nil {
					return "", helpers.ErrorInvalidMethod
				}
				// Prevent out of range error
				if j < 0 {
					j = 0
				} else if j > len(dbEntryData)-1 {
					j = len(dbEntryData) - 1
				}
				// Merge slices (could possibly be done better?) !!!
				entryStart := dbEntryData[:j]
				entryStart = entryStart + cStr
				dbEntryData = entryStart + dbEntryData[j:]
				continue
			}
			return "", helpers.ErrorInvalidMethod
		} else {
			return "", helpers.ErrorInvalidMethodParameters
		}
	}
	return dbEntryData, 0
}

func applyArrayMethods(filter *Filter) int {
	method := filter.methods[0]
	dbEntryPointer := filter.dbEntryData
	dbEntryData := (*dbEntryPointer).([]interface{})
	if item, ok := filter.item.([]interface{}); ok {
		// Basic array methods
		switch method {
		case MethodAppend:
			filter.item = append(dbEntryData, item...)
			return 0

		case MethodPrepend:
			filter.item = append(item, dbEntryData...)
			return 0

		case MethodDelete:
			// Item numbers to delete must be in order of greatest to least
			var lastNum int = len(dbEntryData)
			for _, numb := range item {
				if cNumb, ok := makeInt(numb); ok {
					i := int(cNumb)
					if i >= lastNum {
						return helpers.ErrorInvalidMethodParameters
					} else if i >= 0 {
						dbEntryData = append(dbEntryData[:i], dbEntryData[i+1:]...)
					}
					lastNum = i
				} else {
					return helpers.ErrorInvalidMethodParameters
				}
			}
			filter.item = dbEntryData
			return 0
		}

		// Check for append at index method
		if len(method) >= 10 && method[:8] == MethodAppendAt && method[len(method)-1:len(method)] == MethodAppendAtFin {
			// Convert the text inside brackets to int
			i, iErr := strconv.Atoi(method[8 : len(method)-1])
			if iErr != nil {
				return helpers.ErrorInvalidMethod
			}
			// Prevent out of range error
			if i < 0 {
				i = 0
			} else if i > len(dbEntryData)-1 {
				i = len(dbEntryData) - 1
			}
			// Merge slices (could possibly be done better?) !!!
			entryStart := append([]interface{}{}, dbEntryData[:i]...)
			entryStart = append(entryStart, item...)
			filter.item = append(entryStart, dbEntryData[i:]...)
			return 0
		}
	}

	// Try to convert methods[0] to int for index method
	i, iErr := strconv.Atoi(method)
	if iErr != nil {
		return helpers.ErrorInvalidMethod
	}
	// Prevent out of range error
	if len(dbEntryData) == 0 {
		return helpers.ErrorArrayEmpty
	} else if i < 0 {
		i = 0
	} else if i > len(dbEntryData)-1 {
		i = len(dbEntryData) - 1
	}
	// Check for more methods
	if len(filter.methods) == 1 {
		// No other methods, change value of item
		dbEntryData[i] = filter.item
		filter.item = dbEntryData
		return 0
	} else {
		// More methods to run on item
		filter.dbEntryData = &dbEntryData[i]
		filter.methods = filter.methods[1:]
		filter.schemaItems = append(filter.schemaItems, filter.schemaItems[filter.schemaItemOn].iType.(ArrayItem).dataType.(*SchemaItem))
		filter.schemaItemOn++
		iTypeErr := QueryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		filter.dbEntryData = dbEntryPointer
		filter.item = dbEntryData
		return 0
	}

	return helpers.ErrorInvalidMethod
}

func applyMapMethods(filter *Filter) int {
	method := filter.methods[0]
	dbEntryPointer := filter.dbEntryData
	dbEntryData := (*dbEntryPointer).(map[string]interface{})
	// Delete - eg: ["Mary", "Joe", "Vokome"]
	if item, ok := filter.item.([]interface{}); ok && method == MethodDelete {
		// Delete method
		for _, n := range item {
			if itemName, ok := n.(string); ok {
				delete(dbEntryData, itemName)
			} else {
				return helpers.ErrorInvalidMethodParameters
			}
		}
		filter.item = dbEntryData
		return 0
	} else if item, ok := filter.item.(map[string]interface{}); ok && method == MethodAppend {
		// Append method - eg: {"x": 27, "y": 43}
		for itemName, i := range item {
			dbEntryData[itemName] = i
		}
		filter.item = dbEntryData
		return 0
	}

	// Checking for item with the name method[0] (Items with * not accepted)
	if !strings.Contains(method, "*") {
		if len(filter.methods) == 1 {
			// Add item to, or set item in map
			dbEntryData[method] = filter.item
			filter.item = dbEntryData
			return 0
		} else {
			// More methods to run on item
			data := dbEntryData[method]
			filter.dbEntryData = &data
			filter.methods = filter.methods[1:]
			filter.schemaItems = append(filter.schemaItems, filter.schemaItems[filter.schemaItemOn].iType.(MapItem).dataType.(*SchemaItem))
			filter.schemaItemOn++
			iTypeErr := QueryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			dbEntryData[method] = filter.item
			filter.dbEntryData = dbEntryPointer
			filter.item = dbEntryData
			return 0
		}
	}

	return helpers.ErrorInvalidMethod
}

func applyObjectMethods(filter *Filter) int {
	method := filter.methods[0]
	dbEntryPointer := filter.dbEntryData
	dbEntryData := (*dbEntryPointer).(map[string]interface{})
	si := (*(filter.schemaItems[filter.schemaItemOn].iType.(ObjectItem).schema))[method]
	if si == nil {
		return helpers.ErrorInvalidMethod
	}

	if len(filter.methods) == 1 {
		// No other methods, change value of item
		dbEntryData[method] = filter.item
		filter.item = dbEntryData
		return 0
	} else {
		// More methods to run on item
		data := dbEntryData[method]
		filter.dbEntryData = &data
		filter.methods = filter.methods[1:]
		filter.schemaItems = append(filter.schemaItems, si)
		filter.schemaItemOn++
		iTypeErr := QueryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		dbEntryData[method] = filter.item
		filter.dbEntryData = dbEntryPointer
		filter.item = dbEntryData
		return 0
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ITEM TYPE FILTERS   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func boolFilter(filter *Filter) int {
	if i, ok := filter.item.(bool); ok {
		if filter.schemaItemOn == 0 {
			(*(*filter).dbEntryData) = i
		} else {
			filter.item = i
			filter.schemaItemOn--
		}
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func int8Filter(filter *Filter) int {
	var ic int8
	if i, ok := makeFloat(filter.item); ok {
		ic = int8(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(int8)))
		if mErr != 0 {
			return mErr
		}
		ic = int8(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Int8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func int16Filter(filter *Filter) int {
	var ic int16
	if i, ok := makeFloat(filter.item); ok {
		ic = int16(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(int16)))
		if mErr != 0 {
			return mErr
		}
		ic = int16(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Int16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func int32Filter(filter *Filter) int {
	var ic int32
	if i, ok := makeFloat(filter.item); ok {
		ic = int32(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(int32)))
		if mErr != 0 {
			return mErr
		}
		ic = int32(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Int32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func int64Filter(filter *Filter) int {
	var ic int64
	if i, ok := makeFloat(filter.item); ok {
		ic = int64(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(int64)))
		if mErr != 0 {
			return mErr
		}
		ic = int64(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Int64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func uint8Filter(filter *Filter) int {
	var ic uint8
	if i, ok := makeFloat(filter.item); ok {
		ic = uint8(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(uint8)))
		if mErr != 0 {
			return mErr
		}
		ic = uint8(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Uint8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func uint16Filter(filter *Filter) int {
	var ic uint16
	if i, ok := makeFloat(filter.item); ok {
		ic = uint16(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(uint16)))
		if mErr != 0 {
			return mErr
		}
		ic = uint16(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Uint16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func uint32Filter(filter *Filter) int {
	var ic uint32
	if i, ok := makeFloat(filter.item); ok {
		ic = uint32(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(uint32)))
		if mErr != 0 {
			return mErr
		}
		ic = uint32(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Uint32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func uint64Filter(filter *Filter) int {
	var ic uint64
	if i, ok := makeFloat(filter.item); ok {
		ic = uint64(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(uint64)))
		if mErr != 0 {
			return mErr
		}
		ic = uint64(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Uint64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func float32Filter(filter *Filter) int {
	var ic float32
	if i, ok := makeFloat(filter.item); ok {
		ic = float32(i)
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, float64((*(*filter).dbEntryData).(float32)))
		if mErr != 0 {
			return mErr
		}
		ic = float32(mRes)
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Float32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func float64Filter(filter *Filter) int {
	var ic float64
	if i, ok := makeFloat(filter.item); ok {
		ic = i
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply arithmetic methods
		mRes, mErr := applyNumberMethods(i, filter.methods, (*(*filter).dbEntryData).(float64))
		if mErr != 0 {
			return mErr
		}
		ic = mRes
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(Float64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic * (-1)
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func stringFilter(filter *Filter) int {
	var ic string
	if i, ok := filter.item.(string); ok {
		ic = i
	} else if i, ok := filter.item.([]interface{}); ok && len(filter.methods) > 0 {
		// Apply string methods
		mRes, mErr := applyStringMethods(i, filter.methods, (*(*filter).dbEntryData).(string))
		if mErr != 0 {
			return mErr
		}
		ic = mRes
	} else {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[filter.schemaItemOn].iType.(StringItem)
	l := uint32(len(ic))
	// Check length and if required
	if it.maxChars > 0 && l > it.maxChars {
		return helpers.ErrorStringTooLarge
	} else if it.required && l == 0 {
		return helpers.ErrorStringRequired
	}
	// Check if unique
	if it.unique {
		if filter.schemaItemOn == 0 {
			// Table check
			name := filter.schemaItems[filter.schemaItemOn].name
			(*(*filter).uMux).Lock()
			if (*(*filter).uniqueVals)[name] == nil {
				(*(*filter).uniqueVals)[name] = make(map[string]bool)
			} else if (*(*filter).uniqueVals)[name][ic] {
				(*(*filter).uMux).Unlock()
				return helpers.ErrorUniqueValueInUse
			}
			(*(*filter).uniqueVals)[name][ic] = true
			(*(*filter).uMux).Unlock()
		}

		// Distributed checks !!!!
	}
	if filter.schemaItemOn == 0 {
		(*(*filter).dbEntryData) = ic
	} else {
		filter.item = ic
		filter.schemaItemOn--
	}
	return 0
}

func arrayFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		mErr := applyArrayMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if i, ok := filter.item.([]interface{}); ok {
		filter.methods = []string{}
		it := filter.schemaItems[filter.schemaItemOn].iType.(ArrayItem)
		var iTypeErr int
		// Check inner item type
		for k := 0; k < len(i); k++ {
			filter.item = i[k]
			filter.schemaItems = append(filter.schemaItems, it.dataType.(*SchemaItem))
			filter.schemaItemOn++
			iTypeErr = QueryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[k] = filter.item
		}
		if it.required && len(i) == 0 {
			return helpers.ErrorArrayItemsRequired
		}
		if filter.schemaItemOn == 0 {
			(*(*filter).dbEntryData) = i
		} else {
			filter.item = i
			filter.schemaItemOn--
		}
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func mapFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		mErr := applyMapMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if i, ok := filter.item.(map[string]interface{}); ok {
		filter.methods = []string{}
		it := filter.schemaItems[filter.schemaItemOn].iType.(MapItem)
		var iTypeErr int
		// Check inner item type
		for itemName, _ := range i {
			filter.item = i[itemName]
			filter.schemaItems = append(filter.schemaItems, it.dataType.(*SchemaItem))
			filter.schemaItemOn++
			iTypeErr = QueryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[itemName] = filter.item
		}
		if it.required && len(i) == 0 {
			return helpers.ErrorMapItemsRequired
		}
		if filter.schemaItemOn == 0 {
			(*(*filter).dbEntryData) = i
		} else {
			filter.item = i
			filter.schemaItemOn--
		}
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func objectFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		var mErr int
		mErr = applyObjectMethods(filter)
		if mErr != 0 {
			return mErr
		}
	}
	if i, ok := filter.item.(map[string]interface{}); ok {
		filter.methods = []string{}
		it := filter.schemaItems[filter.schemaItemOn].iType.(ObjectItem)
		filter.schemaItems = append(filter.schemaItems, nil)
		for itemName, schemaItem := range *(it.schema) {
			filter.schemaItemOn++
			filter.schemaItems[filter.schemaItemOn] = schemaItem
			filter.item = i[itemName]
			filterErr := QueryItemFilter(filter)
			if filterErr != 0 {
				return filterErr
			}
			i[itemName] = filter.item
		}
		if filter.schemaItemOn == 0 {
			(*(*filter).dbEntryData) = i
		} else {
			filter.item = i
			filter.schemaItemOn--
		}
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func timeFilter(filter *Filter) int {
	if i, ok := filter.item.(string); ok {
		if i == "*now" {
			// Set to current database time
			if filter.schemaItemOn == 0 {
				(*(*filter).dbEntryData) = time.Now()
			} else {
				filter.item = time.Now()
				filter.schemaItemOn--
			}
			return 0
		}
		it := filter.schemaItems[filter.schemaItemOn].iType.(TimeItem)
		t, tErr := time.Parse(it.format, i)
		if tErr != nil {
			return helpers.ErrorInvalidTimeFormat
		}
		if filter.schemaItemOn == 0 {
			(*(*filter).dbEntryData) = t
		} else {
			filter.item = t
			filter.schemaItemOn--
		}
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ITEM TYPE CONVERTERS   /////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func makeFloat(insertItem interface{}) (float64, bool) {
	switch t := insertItem.(type) {
	case float64:
		return t, true

	case int:
		return float64(t), true

	case int8:
		return float64(t), true

	case int16:
		return float64(t), true

	case int32:
		return float64(t), true

	case int64:
		return float64(t), true

	case uint8:
		return float64(t), true

	case uint16:
		return float64(t), true

	case uint32:
		return float64(t), true

	case uint64:
		return float64(t), true

	case float32:
		return float64(t), true
	}
	return 0, false
}

func makeInt(insertItem interface{}) (int, bool) {
	switch t := insertItem.(type) {
	case int:
		return t, true

	case int8:
		return int(t), true

	case int16:
		return int(t), true

	case int32:
		return int(t), true

	case int64:
		return int(t), true

	case uint8:
		return int(t), true

	case uint16:
		return int(t), true

	case uint32:
		return int(t), true

	case uint64:
		return int(t), true

	case float32:
		return int(t), true

	case float64:
		return int(t), true
	}
	return 0, false
}
