package schema

import (
	"github.com/hewiefreeman/GopherGameDB/helpers"
	"strings"
	"strconv"
	//"fmt"
)

// Arithmetic operators
const (
	OperatorAdd = "+"
	OperatorSub = "-"
	OperatorMul = "*"
	OperatorDiv = "/"
	OperatorMod = "%"
)

// Method names
const (
	MethodOperatorAdd = "*add"
	MethodOperatorSub = "*sub"
	MethodOperatorMul = "*mul"
	MethodOperatorDiv = "*div"
	MethodOperatorMod = "*mod"
	MethodAppend      = "*append"
	MethodPrepend     = "*prepend"
	MethodDelete      = "*delete"
)

// Item type query filters - Initialized when the first Schema is made (see New())
var (
	queryFilters map[string]func(interface{}, []string, interface{}, *SchemaItem)(interface{}, int)
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   QUERY FILTER   /////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// QueryItemFilter takes in an item from a query, and filters/checks it for format/completion against the cooresponding SchemaItem data type.
func QueryItemFilter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if insertItem == nil {
		// Get default value
		defaultVal, defaultErr := defaultVal(itemType)
		if defaultErr != 0 {
			return nil, defaultErr
		}
		return defaultVal, 0
	} else {
		var iTypeErr int
		insertItem, iTypeErr = filterItemType(insertItem, itemMethods, dbEntryData, itemType)
		if iTypeErr != 0 {
			return nil, iTypeErr
		}
		return insertItem, 0
	}
}

func filterItemType(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	return queryFilters[itemType.typeName](insertItem, itemMethods, dbEntryData, itemType)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   QUERY ARITHMETIC   /////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func applyArithmetic(updateItem []interface{}, dbEntryData float64) (float64, int) {
	// Check format & get operator & number for math
	op, num, aErr := checkArithmeticFormat(updateItem)
	if aErr != 0 {
		return 0, aErr
	}
	// Apply arithmetic
	switch op {
		case OperatorAdd:
			return dbEntryData + num, 0

		case OperatorSub:
			return dbEntryData - num, 0

		case OperatorMul:
			return dbEntryData * num, 0

		case OperatorDiv:
			return dbEntryData / num, 0

		case OperatorMod:
			return float64(int(dbEntryData) % int(num)), 0
	}
	return 0, helpers.ErrorInvalidArithmeticOperator
}

func checkArithmeticFormat(updateItem []interface{}) (string, float64, int) {
	// Check format
	if len(updateItem) != 2 {
		return "", 0, helpers.ErrorInvalidArithmeticParameter
	}
	// Get operator
	var ok bool
	var op string
	if op, ok = updateItem[0].(string); !ok {
		return "", 0, helpers.ErrorInvalidArithmeticParameter
	}
	// Get number
	var num float64
	if num, ok = updateItem[1].(float64); !ok {
		return "", 0, helpers.ErrorInvalidArithmeticParameter
	}
	return op, num, 0
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

func applyNumberMethod(numbs []interface{}, methods []string, dbEntryData float64) (float64, int) {
	// Must have same amount of numbers in array as methods to use on them
	if len(numbs) != len(methods) {
		return 0, helpers.ErrorInvalidMethodParameters
	}
	for i, numb := range numbs {
		// Check numb type
		if cNumb, ok := numb.(float64); ok {
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

func applyArrayMethod(insertItem []interface{}, methods []string, dbEntryData []interface{}, itemType *SchemaItem) ([]interface{}, int) {
	// Basic array methods
	switch methods[0] {
		case MethodAppend:
			return append(dbEntryData, insertItem...), 0

		case MethodPrepend:
			return append(insertItem, dbEntryData...), 0

		case MethodDelete:
			// Item numbers to delete must be in order of greatest to least
			var lastNum int = len(dbEntryData)
			for _, numb := range insertItem {
				if cNumb, ok := makeInt(numb); ok {
					i := int(cNumb)
					if i >= lastNum {
						return nil, helpers.ErrorInvalidMethodParameters
					} else if i >= 0 {
						dbEntryData = append(dbEntryData[:i], dbEntryData[i+1:]...)
					}
					lastNum = i
				} else {
					return nil, helpers.ErrorInvalidMethodParameters
				}
			}
			return dbEntryData, 0
	}

	// Check for append at index method
	if len(methods[0]) >= 10 && methods[0][:8] == "*append[" && methods[0][len(methods[0])-1:len(methods[0])] == "]" {
		// Convert the text inside brackets to int
		i, iErr := strconv.Atoi(methods[0][8:len(methods[0])-1])
		if iErr != nil {
			return nil, helpers.ErrorInvalidMethod
		}

		if i < 0 {
			i = 0
		} else if i > len(dbEntryData)-1 {
			i = len(dbEntryData)-1
		}

		entryStart := append([]interface{}{}, dbEntryData[:i]...)
		entryStart = append(entryStart, insertItem...)
		return append(entryStart, dbEntryData[i:]...), 0
	}

	// Try to convert methods[0] to int for index method
	/*i, iErr := strconv.Atoi(methods[0])
	if iErr != nil {
		return 0, helpers.ErrorInvalidMethod
	}*/

	return nil, helpers.ErrorInvalidMethod
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//   ITEM TYPE FILTERS   ////////////////////////////////////////////////////////////////////////////////////////
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

func boolFilter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.(bool); ok {
		return i, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}

func int8Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int8
	if i, ok := makeFloat(insertItem); ok {
		ic = int8(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(int8)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = int8(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(int8)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = int8(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func int16Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int16
	if i, ok := makeFloat(insertItem); ok {
		ic = int16(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(int16)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = int16(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(int16)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = int16(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func int32Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int32
	if i, ok := makeFloat(insertItem); ok {
		ic = int32(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(int32)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = int32(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(int32)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = int32(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func int64Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic int64
	if i, ok := makeFloat(insertItem); ok {
		ic = int64(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(int64)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = int64(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(int64)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = int64(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Int64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint8Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint8
	if i, ok := makeFloat(insertItem); ok {
		ic = uint8(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(uint8)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = uint8(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(uint8)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = uint8(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint8Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint16Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint16
	if i, ok := makeFloat(insertItem); ok {
		ic = uint16(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(uint16)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = uint16(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(uint16)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = uint16(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint16Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint32Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint32
	if i, ok := makeFloat(insertItem); ok {
		ic = uint32(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(uint32)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = uint32(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(uint32)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = uint32(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func uint64Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic uint64
	if i, ok := makeFloat(insertItem); ok {
		ic = uint64(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(uint64)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = uint64(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(uint64)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = uint64(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Uint64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	return ic, 0
}

func float32Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic float32
	if i, ok := makeFloat(insertItem); ok {
		ic = float32(i)
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			mRes, mErr := applyNumberMethod(i, itemMethods, float64(dbEntryData.(float32)))
			if mErr != 0 {
				return nil, mErr
			}
			ic = float32(mRes)
		} else {
			aRes, aErr := applyArithmetic(i, float64(dbEntryData.(float32)))
			if aErr != 0 {
				return nil, aErr
			}
			ic = float32(aRes)
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Float32Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic*(-1)
	}
	return ic, 0
}

func float64Filter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	var ic float64
	if i, ok := makeFloat(insertItem); ok {
		ic = i
	} else if i, ok := insertItem.([]interface{}); ok {
		// Apply method or arithmetic if no methods provided
		if len(itemMethods) > 0 {
			var mErr int
			ic, mErr = applyNumberMethod(i, itemMethods, dbEntryData.(float64))
			if mErr != 0 {
				return nil, mErr
			}
		} else {
			var aErr int
			ic, aErr = applyArithmetic(i, dbEntryData.(float64))
			if aErr != 0 {
				return nil, aErr
			}
		}
	} else {
		return nil, helpers.ErrorInvalidItemValue
	}
	it := itemType.iType.(Float64Item)
	// Check min/max unless both are the same
	if it.min < it.max {
		if ic > it.max {
			ic = it.max
		} else if ic < it.min {
			ic = it.min
		}
	}
	if it.abs && ic < 0 {
		ic = ic*(-1)
	}
	return ic, 0
}

func stringFilter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.(string); ok {
		it := itemType.iType.(StringItem)
		l := uint32(len(i))
		// Check length and if required
		if it.maxChars > 0 && l > it.maxChars {
			return nil, helpers.ErrorStringTooLarge
		} else if it.required && l == 0 {
			return nil, helpers.ErrorStringRequired
		}
		// Check if unique
		if it.unique {
			// unique checks !!!!!!
		}
		return i, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}

func arrayFilter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.([]interface{}); ok {
		if len(itemMethods) > 0 {
			var mErr int
			i, mErr = applyArrayMethod(i, itemMethods, dbEntryData.([]interface{}), itemType)
			if mErr != 0 {
				return nil, mErr
			}
		}
		it := itemType.iType.(ArrayItem)
		var iTypeErr int
		// Check inner item type
		for k := 0; k < len(i); k++ {
			i[k], iTypeErr = filterItemType(i[k], nil, nil, it.dataType.(*SchemaItem))
			if iTypeErr != 0 {
				return nil, iTypeErr
			}
		}
		return i, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}

func objectFilter(insertItem interface{}, itemMethods []string, dbEntryData interface{}, itemType *SchemaItem) (interface{}, int) {
	if i, ok := insertItem.(map[string]interface{}); ok {
		it := itemType.iType.(ObjectItem)
		newObj := make(map[string]interface{})
		var filterErr int
		for itemName, schemaItem := range *(it.schema) {
			innerItem := i[itemName]
			newObj[itemName], filterErr = QueryItemFilter(innerItem, nil, nil, schemaItem)
			if filterErr != 0 {
				return nil, filterErr
			}
		}
		return newObj, 0
	}
	return nil, helpers.ErrorInvalidItemValue
}