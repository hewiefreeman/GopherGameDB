package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"time"
)

// Filter for queries
type Filter struct {
	restore     bool
	get         bool                    // when true, output is for get queries
	eCost       int                     // encryption cost for encrypted items
	item        interface{}             // The item data to insert/get, or method parameters when methods are being used
	destination *interface{}            // Pointer to where the filtered data must go
	methods     []string                // Method list
	innerData   []interface{}           // Data hierarchy holder for entry on database - used for unique value searches and methods
	schemaItems []SchemaItem            // Schema hierarchy holder - used for unique value searches
	uniqueVals  *map[string]interface{} // Pointer a map storing all unique values to check against their table after running filter
}

// ItemFilter filters an item in a query against it's corresponding SchemaItem.
func ItemFilter(item interface{}, methods []string, destination *interface{}, innerData interface{}, schemaItem SchemaItem, uniqueVals *map[string]interface{}, eCost int, get bool, restore bool) int {
	filter := Filter{
		restore:     restore,
		get:         get,
		eCost:       eCost,
		item:        item,
		methods:     methods,
		destination: destination,
		innerData:   []interface{}{},
		schemaItems: []SchemaItem{schemaItem},
		uniqueVals:  uniqueVals,
	}
	if innerData != nil {
		filter.innerData = []interface{}{innerData}
	}
	return queryItemFilter(&filter)
}

// queryItemFilter takes in an item from a query, and filters/checks it for format/completion against the corresponding SchemaItem data type.
func queryItemFilter(filter *Filter) int {
	if !filter.get && filter.item == nil {
		// No methods allowed on a nil item
		if len(filter.methods) > 0 {
			return helpers.ErrorInvalidMethodParameters
		}
		// Get default value
		dVal, defaultErr := defaultVal(filter.schemaItems[len(filter.schemaItems)-1])
		if defaultErr != 0 {
			return defaultErr
		}
		if len(filter.schemaItems) == 1 {
			(*(*filter).destination) = dVal
		} else {
			filter.item = dVal
		}
		return 0
	}

	// Run type filter
	iTypeErr := getTypeFilter(filter.schemaItems[len(filter.schemaItems)-1].typeName)(filter)
	if iTypeErr != 0 {
		return iTypeErr
	}
	// Check if this is the last filter itteration, and apply item to destination
	if len(filter.schemaItems) == 1 {
		(*(*filter).destination) = filter.item
	}
	return 0
}

func getTypeFilter(typeName string) func(*Filter) int {
	switch typeName {
	case ItemTypeBool:
		return boolFilter
	case ItemTypeInt8:
		return int8Filter
	case ItemTypeInt16:
		return int16Filter
	case ItemTypeInt32:
		return int32Filter
	case ItemTypeInt64:
		return int64Filter
	case ItemTypeUint8:
		return uint8Filter
	case ItemTypeUint16:
		return uint16Filter
	case ItemTypeUint32:
		return uint32Filter
	case ItemTypeUint64:
		return uint64Filter
	case ItemTypeFloat32:
		return float32Filter
	case ItemTypeFloat64:
		return float64Filter
	case ItemTypeString:
		return stringFilter
	case ItemTypeArray:
		return arrayFilter
	case ItemTypeMap:
		return mapFilter
	case ItemTypeObject:
		return objectFilter
	case ItemTypeTime:
		return timeFilter
	default:
		return nil
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   Item type filters   ////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func boolFilter(filter *Filter) int {
	if filter.get {
		filter.methods = []string{}
		filter.item = filter.innerData[len(filter.innerData)-1]
		return 0
	} else if i, ok := filter.item.(bool); ok {
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func int8Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyIntMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic int8
	var ok bool
	if ic, ok = makeInt8(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int8Item)
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
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func int16Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyIntMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic int16
	var ok bool
	if ic, ok = makeInt16(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int16Item)
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
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func int32Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyIntMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic int32
	var ok bool
	if ic, ok = makeInt32(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int32Item)
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
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func int64Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyIntMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic int64
	var ok bool
	if ic, ok = makeInt64(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Int64Item)
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
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	filter.item, _ = makeTypeStorage(filter.item, &filter.schemaItems[len(filter.schemaItems)-1])
	return 0
}

func uint8Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyUintMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic uint8
	var ok bool
	if ic, ok = makeUint8(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func uint16Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyUintMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic uint16
	var ok bool
	if ic, ok = makeUint16(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func uint32Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyUintMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic uint32
	var ok bool
	if ic, ok = makeUint32(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func uint64Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyUintMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic uint64
	var ok bool
	if ic, ok = makeUint64(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Uint64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	filter.item, _ = makeTypeStorage(filter.item, &filter.schemaItems[len(filter.schemaItems)-1])
	return 0
}

func float32Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyFloatMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item, _ = makeTypeLiteral(filter.innerData[len(filter.innerData)-1], &filter.schemaItems[len(filter.schemaItems)-1])
		return 0
	}
	var ic float32
	var ok bool
	if ic, ok = makeFloat32(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Float32Item)
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
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func float64Filter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply number methods
		mErr := applyFloatMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		filter.item = filter.innerData[len(filter.innerData)-1]
		return 0
	}
	var ic float64
	var ok bool
	if ic, ok = makeFloat64(filter.item); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(Float64Item)
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
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func stringFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Apply string methods
		mErr := applyStringMethods(filter)
		if mErr != 0 {
			return mErr
		}
		if filter.get {
			return 0
		}
	} else if filter.get {
		// Cannot retrieve an encrypted item directly
		if filter.schemaItems[len(filter.schemaItems)-1].iType.(StringItem).encrypted {
			return helpers.ErrorStringIsEncrypted
		}
		filter.item = filter.innerData[len(filter.innerData)-1]
		return 0
	}
	var ic string
	var ok bool
	if ic, ok = filter.item.(string); !ok {
		return helpers.ErrorInvalidItemValue
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(StringItem)
	// Don't filter encrypted strings while restoring
	if filter.restore && it.encrypted {
		filter.item = ic
		return 0
	}
	// Check length and if required
	l := uint32(len(ic))
	if it.maxChars > 0 && l > it.maxChars {
		return helpers.ErrorStringTooLarge
	} else if it.required && l == 0 {
		return helpers.ErrorStringRequired
	}
	if it.encrypted {
		// Encrypt ic
		var es []byte
		var err error
		if es, err = helpers.EncryptString(ic, filter.eCost); err != nil {
			return helpers.ErrorEncryptingString
		}
		// Encrypted strings cannot be unique. Skip next checks
		filter.item = string(es)
		return 0
	}
	// Check if unique
	filter.item = ic
	if it.unique && uniqueCheck(filter) {
		return helpers.ErrorUniqueValueDuplicate
	}
	return 0
}

func arrayFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Copy Array to prevent prematurely changing data in entry's pointer to this innerData slice
		filter.innerData[len(filter.innerData)-1] = append([]interface{}{}, filter.innerData[len(filter.innerData)-1].([]interface{})...)
		mErr := applyArrayMethods(filter)
		if mErr != 0 {
			return mErr
		}
		return 0
	} else if filter.get {
		filter.item = filter.innerData[len(filter.innerData)-1]
		return filterArrayGetQuery(filter)
	} else if i, ok := filter.item.([]interface{}); ok {
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem)
		var iTypeErr int
		// Check inner item type
		if (len(filter.schemaItems) == 1 && len(filter.innerData) == 0) || len(filter.schemaItems) > 1 {
			filter.innerData = append(filter.innerData, make([]interface{}, 0))
		}
		filter.schemaItems = append(filter.schemaItems, it.dataType)
		var index int
		for index, filter.item = range i {
			iTypeErr = queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[index] = filter.item
			filter.innerData[len(filter.innerData)-1] = append(filter.innerData[len(filter.innerData)-1].([]interface{}), filter.item)
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems)-1]
		if it.required && len(i) == 0 {
			return helpers.ErrorArrayItemsRequired
		}
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func filterArrayGetQuery(filter *Filter) int {
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem)
	i := append([]interface{}{}, filter.item.([]interface{})...)
	filter.schemaItems = append(filter.schemaItems, it.dataType)
	filter.innerData = append(filter.innerData, nil)
	var index int
	var iTypeErr int
	for index, filter.item = range i {
		filter.innerData[len(filter.innerData)-1] = filter.item
		iTypeErr = queryItemFilter(filter)
		if iTypeErr != 0 {
			return iTypeErr
		}
		i[index] = filter.item
	}
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems)-1]
	filter.innerData = filter.innerData[:len(filter.innerData)-1]
	filter.item = i
	return 0
}

func mapFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		// Copy Map to prevent prematurely changing data in entry's pointer to this innerData map
		var m map[string]interface{} = make(map[string]interface{}, len(filter.innerData[len(filter.innerData)-1].(map[string]interface{})))
		for n, v := range filter.innerData[len(filter.innerData)-1].(map[string]interface{}) {
			m[n] = v
		}
		filter.innerData[len(filter.innerData)-1] = m
		mErr := applyMapMethods(filter)
		if mErr != 0 {
			return mErr
		}
		return 0
	} else if filter.get {
		filter.item = filter.innerData[len(filter.innerData)-1]
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem)
		switch it.dataType.typeName {
		case ItemTypeObject, ItemTypeArray, ItemTypeMap:
			// Copy Map to prevent changing data in entry's pointer to this innerData map
			var m map[string]interface{} = make(map[string]interface{})
			for n, v := range filter.innerData[len(filter.innerData)-1].(map[string]interface{}) {
				m[n] = v
			}
			filter.schemaItems = append(filter.schemaItems, it.dataType)
			filter.innerData = append(filter.innerData, nil)
			var itemName string
			var iTypeErr int
			for itemName, filter.item = range m {
				filter.innerData[len(filter.innerData)-1] = filter.item
				iTypeErr = queryItemFilter(filter)
				if iTypeErr != 0 {
					return iTypeErr
				}
				m[itemName] = filter.item
			}
			filter.schemaItems = filter.schemaItems[:len(filter.schemaItems)-1]
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			filter.item = m
		}
		return 0
	} else if i, ok := filter.item.(map[string]interface{}); ok {
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem)
		// Check inner item type
		if (len(filter.schemaItems) == 1 && len(filter.innerData) == 0) || len(filter.schemaItems) > 1 {
			filter.innerData = append(filter.innerData, make(map[string]interface{}))
		}
		filter.schemaItems = append(filter.schemaItems, it.dataType)
		var itemName string
		var iTypeErr int
		for itemName, filter.item = range i {
			iTypeErr = queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			i[itemName] = filter.item
			filter.innerData[len(filter.innerData)-1].(map[string]interface{})[itemName] = filter.item
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems)-1]
		if it.required && len(i) == 0 {
			return helpers.ErrorMapItemsRequired
		}
		filter.item = i
		return 0
	}
	return helpers.ErrorInvalidItemValue
}

func objectFilter(filter *Filter) int {
	if len(filter.methods) > 0 {
		mErr := applyObjectMethods(filter)
		if mErr != 0 {
			return mErr
		}
		return 0
	} else if filter.get {
		// Convert data to map[string]interface{}
		objList := append([]interface{}{}, filter.innerData[len(filter.innerData)-1].([]interface{})...)
		m := make(map[string]interface{})
		var iTypeErr int
		for itemName, schemaItem := range filter.schemaItems[len(filter.schemaItems)-1].iType.(ObjectItem).schema {
			filter.schemaItems = append(filter.schemaItems, schemaItem)
			filter.innerData = append(filter.innerData, objList[schemaItem.dataIndex])
			iTypeErr = queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			filter.schemaItems = filter.schemaItems[:len(filter.schemaItems)-1]
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
			m[itemName] = filter.item
		}
		filter.item = m
		return 0
	}
	it := filter.schemaItems[len(filter.schemaItems)-1].iType.(ObjectItem)
	if (len(filter.schemaItems) == 1 && len(filter.innerData) == 0) || len(filter.schemaItems) > 1 {
		filter.innerData = append(filter.innerData, make([]interface{}, len(it.schema), len(it.schema)))
	}
	filter.schemaItems = append(filter.schemaItems, SchemaItem{})
	if i, ok := filter.item.(map[string]interface{}); ok {
		// Object format
		for itemName, schemaItem := range it.schema {
			filter.schemaItems[len(filter.schemaItems)-1] = schemaItem
			filter.item = i[itemName]
			filterErr := queryItemFilter(filter)
			if filterErr != 0 {
				return filterErr
			}
			filter.innerData[len(filter.innerData)-1].([]interface{})[schemaItem.dataIndex] = filter.item
		}
		filter.item = filter.innerData[len(filter.innerData)-1]

	} else if i, ok := filter.item.([]interface{}); ok {
		// List format
		for _, schemaItem := range it.schema {
			filter.schemaItems[len(filter.schemaItems)-1] = schemaItem
			// Prevent out of range
			if int(schemaItem.dataIndex) >= len(i) {
				filter.item = nil
			} else {
				filter.item = i[schemaItem.dataIndex]
			}
			filterErr := queryItemFilter(filter)
			if filterErr != 0 {
				return filterErr
			}
			filter.innerData[len(filter.innerData)-1].([]interface{})[schemaItem.dataIndex] = filter.item
		}
		filter.item = filter.innerData[len(filter.innerData)-1]
	} else {
		return helpers.ErrorInvalidItemValue
	}
	filter.innerData = filter.innerData[:len(filter.innerData)-1]
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems)-1]
	return 0
}

func timeFilter(filter *Filter) int {
	if filter.get {
		var t time.Time
		// If the item is a string, was retrieved from disk - convert to time.Time
		if i, ok := filter.innerData[len(filter.innerData)-1].(string); ok {
			var tErr error
			t, tErr = time.Parse(TimeFormatRFC3339, i) // JSON uses RFC3339
			if tErr != nil {
				return helpers.ErrorInvalidTimeFormat
			}
		} else {
			t = filter.innerData[len(filter.innerData)-1].(time.Time)
		}
		if len(filter.methods) > 0 {
			if mErr := applyTimeMethods(filter, t); mErr != 0 {
				return mErr
			}
			return 0
		}
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(TimeItem)
		filter.item = t.Format(it.format)
		return 0
	} else if i, ok := filter.item.(string); ok {
		if len(filter.methods) > 0 {
			return helpers.ErrorInvalidMethod
		}
		if i == "*now" {
			// Set to current database time
			filter.item = time.Now()
			return 0
		}
		it := filter.schemaItems[len(filter.schemaItems)-1].iType.(TimeItem)
		var t time.Time
		var err error
		if filter.restore {
			// Restoring from JSON file using RFC3339
			t, err = time.Parse(TimeFormatRFC3339, i)
		} else {
			t, err = time.Parse(it.format, i)
		}
		if err != nil {
			return helpers.ErrorInvalidTimeFormat
		}
		filter.item = t
		return 0
	}
	return helpers.ErrorInvalidItemValue
}
