package schema

import (
	"github.com/hewiefreeman/GopherDB/helpers"
	"time"
	"strconv"
	"strings"
	//"fmt"
)

/*  **** RULES ****

	- Query methods must always be paired with a parameter ([]interface{}) list, or (map[string]interface{}) map for certain Map methods

*/

// Method names
const (
	MethodOperatorAdd = "*add"
	MethodOperatorSub = "*sub"
	MethodOperatorMul = "*mul"
	MethodOperatorDiv = "*div"
	MethodOperatorMod = "*mod"
	MethodLength      = "*len"
	MethodEquals      = "*eq"
	MethodGreater     = "*gt"
	MethodLess        = "*lt"
	MethodGreaterOE   = "*gte"
	MethodLessOE      = "*lte"
	MethodContains    = "*contains" // For Arrays and Maps
	MethodIndexOf     = "*indexOf"  // For Arrays
	MethodKeyOf       = "*keyOf"    // For Maps
	MethodSortAsc     = "*sort"     // For Arrays | TO-DO
	MethodSortDesc    = "*sortDesc" // For Arrays | TO-DO
	MethodAppend      = "*append"
	MethodAppendAt    = "*append["
	MethodAppendAtFin = "]"
	MethodPrepend     = "*prepend"
	MethodDelete      = "*delete"
	MethodSince       = "*since"
	MethodUntil       = "*until"
	MethodDay         = "*day"
	MethodHour        = "*hr"
	MethodMinute      = "*min"
	MethodSecond      = "*sec"
	MethodMillisecond = "*mil"
	MethodGet         = "*get" // Makes a separate query | TO-DO
	MethodItem        = "*item" // Gets an item from the current entry | TO-DO
)

// GetQueryItemMethods checks query item names for methods and returns the item name and the list of methods.
func GetQueryItemMethods(itemName string) (string, []string) {
	if strings.Contains(itemName, ".") {
		ml := strings.Split(itemName, ".")
		return ml[0], ml[1:]
	}
	return itemName, nil
}

// Run methods on number type item
func applyNumberMethods(filter *Filter) int {
	var entryData float64
	var err int
	var brk bool
	entryData, _ = makeFloat64(filter.innerData[len(filter.innerData)-1])
	if fList, ok := filter.item.([]interface{}); ok {
		for _, methodParam := range fList {
			// Check methodParam type
			if cNumb, ok := makeFloat64(methodParam); ok {
				brk, err = getNumberMethodResult(filter, &entryData, cNumb)
				if err != 0 {
					return err
				}
			} else {
				return helpers.ErrorInvalidMethodParameters
			}
			// Break when requested (when entrydata would no longer be a number type)
			if brk {
				break
			}
			// Remove this method
			filter.methods = filter.methods[1:]
		}
	} else {
		return helpers.ErrorInvalidMethodParameters
	}
	filter.methods = []string{}
	if !brk {
		// Convert filter item to int if this is a get query and the type of number is not float
		if filter.get && !filter.schemaItems[len(filter.schemaItems) - 1].IsFloat() {
			filter.item, _ = makeInt(entryData)
		} else {
			filter.item = entryData
		}
	}
	return 0
}

func getNumberMethodResult(filter *Filter, entryData *float64, num float64) (bool, int) {
	if len(filter.methods) == 0 {
		return false, helpers.ErrorTooManyMethodParameters
	}
	method := filter.methods[0]
	var brk bool
	if filter.get {
		switch method {
			case MethodEquals:
				filter.item = (*entryData == num)
				brk = true

			case MethodGreater:
				filter.item = (*entryData > num)
				brk = true

			case MethodGreaterOE:
				filter.item = (*entryData >= num)
				brk = true

			case MethodLess:
				filter.item = (*entryData < num)
				brk = true

			case MethodLessOE:
				filter.item = (*entryData <= num)
				brk = true

			default:
				if err := checkGeneralNumberMethods(method, entryData, num); err != 0 {
					return false, err
				}
		}
	} else {
		if err := checkGeneralNumberMethods(method, entryData, num); err != 0 {
			return false, err
		}
	}
	return brk, 0
}

func checkGeneralNumberMethods(method string, entryData *float64, num float64) int {
	switch method {
		case MethodOperatorAdd:
			*entryData = *entryData + num

		case MethodOperatorSub:
			*entryData = *entryData - num

		case MethodOperatorMul:
			*entryData = *entryData * num

		case MethodOperatorDiv:
			*entryData = *entryData / num

		case MethodOperatorMod:
			*entryData = float64(int(*entryData + 0.5) % int(num + 0.5))

		default:
			return helpers.ErrorInvalidMethod
	}
	return 0
}

func applyStringMethods(filter *Filter) int {
	var entryData string
	var err int
	var brk bool
	var typeName string
	entryData, _ = filter.innerData[len(filter.innerData)-1].(string)
	if fList, ok := filter.item.([]interface{}); ok {
		for _, methodParam := range fList {
			// Check methodParam type
			if cString, ok := methodParam.(string); ok {
				brk, typeName, err = getStringMethodResult(filter, &entryData, cString)
				if err != 0 {
					return err
				}
			} else {
				brk, typeName, err = getStringMethodResult(filter, &entryData, "")
				if err != 0 {
					return err
				}
			}
			filter.methods = filter.methods[1:]
			if brk {
				break
			}
		}
	} else {
		return helpers.ErrorInvalidMethodParameters
	}
	if brk {
		if len(filter.methods) > 0 {
			// More methods to run...
			switch typeName {
				case ItemTypeFloat64:
					if err = applyNumberMethods(filter); err != 0 {
						return err
					}
			}
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
		} else {
			filter.item = filter.innerData[len(filter.innerData)-1]
			filter.innerData = filter.innerData[:len(filter.innerData)-1]
		}
	} else {
		filter.methods = []string{}
		filter.item = entryData
	}
	return 0
}

func getStringMethodResult(filter *Filter, entryData *string, str string) (bool, string, int) {
	if len(filter.methods) == 0 {
		return false, "", helpers.ErrorTooManyMethodParameters
	}
	method := filter.methods[0]
	var brk bool
	var typeName string
	if filter.get {
		switch method {
			case MethodLength:
				filter.innerData = append(filter.innerData, float64(len(*entryData)))
				typeName = ItemTypeFloat64
				brk = true

			case MethodIndexOf:
				var indexOf float64 = -1
				for i := 0; i < len(*entryData) - (len(str) - 1); i++ {
					if (*entryData)[i:i + len(str)] == str {
						indexOf = float64(i)
						break
					}
				}
				filter.item = filter.item.([]interface{})[1:]
				filter.innerData = append(filter.innerData, indexOf)
				typeName = ItemTypeFloat64
				brk = true

			case MethodContains:
				var contains bool
				for i := 0; i < len(*entryData) - (len(str) - 1); i++ {
					if (*entryData)[i:i + len(str)] == str {
						contains = true
						break
					}
				}
				filter.innerData = append(filter.innerData, contains)
				typeName = ItemTypeBool
				brk = true

			case MethodEquals:
				filter.innerData = append(filter.innerData, (*entryData == str))
				typeName = ItemTypeBool
				brk = true

			default:
				if err := checkGeneralStringMethods(filter, method, entryData, str); err != 0 {
					return false, "", err
				}
		}
	} else {
		if err := checkGeneralStringMethods(filter, method, entryData, str); err != 0 {
			return false, "", err
		}
	}
	return brk, typeName, 0
}

func checkGeneralStringMethods(filter *Filter, method string, entryData *string, str string) int {
	switch method {
		case MethodOperatorAdd, MethodAppend:
			*entryData = *entryData + str

		case MethodPrepend:
			*entryData = str + *entryData

		default:
			// Check for append at index method
			if len(method) >= 10 && method[:8] == MethodAppendAt && method[len(method)-1:len(method)] == MethodAppendAtFin {
				// Convert the text inside brackets to int
				j, jErr := strconv.Atoi(method[8 : len(method)-1])
				if jErr != nil {
					return helpers.ErrorInvalidMethod
				}
				// Prevent out of range error
				if j < 0 {
					j = 0
				} else if j > len(*entryData) - 1 {
					j = len(*entryData) - 1
				}
				// Merge slices (could possibly be done better?) !!!
				entryStart := (*entryData)[:j]
				entryStart = entryStart + str
				*entryData = entryStart + (*entryData)[j:]
			} else {
				return helpers.ErrorInvalidMethod
			}
	}
	filter.item = filter.item.([]interface{})[1:]
	return 0
}

// Run methods on Array item collection
func applyArrayMethods(filter *Filter) int {
	if filter.item == nil {
		return helpers.ErrorInvalidMethodParameters
	}
	method := filter.methods[0]
	dbEntryData := filter.innerData[len(filter.innerData)-1].([]interface{})
	if item, ok := filter.item.([]interface{}); ok {
		if filter.get {
			// Check get query array methods
			switch method {
			case MethodLength:
				filter.methods = filter.methods[1:]
				if len(filter.methods) > 0 {
					// run methods as float64
					filter.innerData = append(filter.innerData, float64(len(dbEntryData)))
					if err := applyNumberMethods(filter); err != 0 {
						return err
					}
					filter.innerData = filter.innerData[:len(filter.innerData) - 1]
				} else {
					filter.item = len(dbEntryData)
				}
				return 0

			case MethodIndexOf:
				if len(item) == 0 {
					return helpers.ErrorInvalidMethodParameters
				}
				var indexOf float64
				var err int
				if indexOf, err = arrayIndexOf(filter, item[0], dbEntryData); err != 0 {
					return err
				}
				if len(filter.methods) > 0 {
					// run methods as float64
					filter.innerData = append(filter.innerData, indexOf)
					if err = applyNumberMethods(filter); err != 0 {
						return err
					}
					filter.innerData = filter.innerData[:len(filter.innerData) - 1]
				} else {
					filter.item = indexOf
				}
				return 0

			case MethodContains:
				if len(item) == 0 {
					return helpers.ErrorInvalidMethodParameters
				}
				var indexOf float64
				var err int
				if indexOf, err = arrayIndexOf(filter, item[0], dbEntryData); err != 0 {
					return err
				}
				filter.methods = []string{}
				filter.item = (indexOf != -1)
				return 0
			}
		} else {
			// Update query array methods
			switch method {
			case MethodAppend:
				// Filter items
				if err := filterArrayAppendMethodItems(filter, item); err != 0 {
					return err
				}
				// Append data
				dbEntryData = append(dbEntryData, item[0].([]interface{})...)
				// Check for more methods
				if len(filter.methods) > 0 {
					filter.item = item[1:]
					filter.innerData[len(filter.innerData) - 1] = dbEntryData
					if err := applyArrayMethods(filter); err != 0 {
						return err
					}
					return 0
				}
				filter.item = dbEntryData
				return 0

			case MethodPrepend:
				// Filter items
				if err := filterArrayAppendMethodItems(filter, item); err != 0 {
					return err
				}
				// Prepend data
				dbEntryData = append(item[0].([]interface{}), dbEntryData...)
				// Check for more methods
				if len(filter.methods) > 0 {
					filter.item = item[1:]
					filter.innerData[len(filter.innerData) - 1] = dbEntryData
					if err := applyArrayMethods(filter); err != 0 {
						return err
					}
					return 0
				}
				filter.item = dbEntryData
				return 0

			case MethodDelete:
				if len(item) == 0 {
					return helpers.ErrorNotEnoughMethodParameters
				}
				// Get delete params
				if mParams, ok := item[0].([]interface{}); ok {
					// Item numbers to delete must be in order of greatest to least
					var lastNum int = len(dbEntryData)
					for _, numb := range mParams {
						if i, ok := makeInt(numb); ok {
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
				}
				filter.methods = filter.methods[1:]
				// Check for more methods
				if len(filter.methods) > 0 {
					filter.item = item[1:]
					filter.innerData[len(filter.innerData) - 1] = dbEntryData
					if err := applyArrayMethods(filter); err != 0 {
						return err
					}
					return 0
				}
				filter.item = dbEntryData
				return 0

			case MethodSortAsc:
				filter.item = dbEntryData
				return 0
			case MethodSortDesc:
				filter.item = dbEntryData
				return 0
			}

			// Check for append at index method
			mLen := len(method)
			if mLen >= 10 && method[:8] == MethodAppendAt && method[mLen-1:mLen] == MethodAppendAtFin {
				// Convert the text inside brackets to int
				i, iErr := strconv.Atoi(method[8 : mLen-1])
				if iErr != nil {
					return helpers.ErrorInvalidMethod
				}
				// Prevent out of range error
				if i < 0 {
					i = 0
				} else if i > len(dbEntryData)-1 {
					i = len(dbEntryData) - 1
				}
				if err := filterArrayAppendMethodItems(filter, item); err != 0 {
					return err
				}
				// Merge slices (could possibly be done better? TO-DO ?)
				entryStart := append([]interface{}{}, dbEntryData[:i]...)
				entryStart = append(entryStart, item[0].([]interface{})...)
				dbEntryData = append(entryStart, dbEntryData[i:]...)
				// Check for more methods
				if len(filter.methods) > 0 {
					filter.item = item[1:]
					filter.innerData[len(filter.innerData) - 1] = dbEntryData
					if err := applyArrayMethods(filter); err != 0 {
						return err
					}
					return 0
				}
				filter.item = dbEntryData
				return 0
			}
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
	} else if i < 0 || i > len(dbEntryData) - 1 {
		return helpers.ErrorIndexOutOfBounds
	}
	// Check for more methods & filter
	filter.methods = filter.methods[1:]
	filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems) - 1].iType.(ArrayItem).dataType)
	filter.innerData = append(filter.innerData, dbEntryData[i])
	iTypeErr := queryItemFilter(filter)
	if iTypeErr != 0 {
		return iTypeErr
	}
	filter.innerData = filter.innerData[:len(filter.innerData) - 1]
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
	if !filter.get {
		dbEntryData[i] = filter.item
		filter.item = dbEntryData
	}
	return 0
}

/*func sortArray(ary []interface{}, itemType SchemaItem, asc bool) int {
	if itemType.IsNumeric() {
		var t float64
		for i := 0; i < len(sa) - 1; i++ {
			for j := len(sa) - 1; j > i; j-- {
				if sa[i] > sa[j] {
					t = sa[i]
					sa[i] = sa[j]
					sa[j] = t
				}
			}
		}
	} else if itemType.typeName == ItemTypeString {

	}
	return helpers.ErrorArrayNotSortable
}*/

func arrayIndexOf(filter *Filter, searchItem interface{}, dbEntryData []interface{}) (float64, int) {
	// Get inner data type
	si := filter.schemaItems[len(filter.schemaItems) - 1].iType.(ArrayItem).dataType
	var indexOf float64 = -1
	if si.IsNumeric() {
		var ok bool
		if searchItem, ok = makeType(searchItem, si); !ok {
			return 0, helpers.ErrorInvalidMethodParameters
		}
		for i, innerItem := range dbEntryData {
			if innerItem, ok = makeType(innerItem, si); !ok {
				return 0, helpers.ErrorUnexpected
			}
			if searchItem == innerItem {
				indexOf = float64(i)
				break
			}
		}
	} else if si.typeName == ItemTypeString || si.typeName == ItemTypeBool {
		for i, innerItem := range dbEntryData {
			if searchItem == innerItem {
				indexOf = float64(i)
				break
			}
		}
	} else {
		return 0, helpers.ErrorInvalidMethod
	}
	filter.item = filter.item.([]interface{})[1:]
	filter.methods = filter.methods[1:]
	return indexOf, 0
}

// Run filter on Array append method item collection
func filterArrayAppendMethodItems(filter *Filter, item []interface{}) int {
	if len(item) == 0 {
		return helpers.ErrorNotEnoughMethodParameters
	}
	if mParams, ok := item[0].([]interface{}); ok {
		// Disallow methods on append items
		m := append([]string{}, filter.methods[1:]...)
		filter.methods = []string{}
		filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(ArrayItem).dataType)
		var index int
		for index, filter.item = range mParams {
			iTypeErr := queryItemFilter(filter)
			if iTypeErr != 0 {
				return iTypeErr
			}
			mParams[index] = filter.item
			// Add item to array to check for duplicate unique values
			filter.innerData[len(filter.innerData)-1] = append(filter.innerData[len(filter.innerData)-1].([]interface{}), filter.item)
		}
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		filter.methods = m
		item[0] = mParams
		return 0
	}
	return helpers.ErrorInvalidMethodParameters
}

// Run methods on Map item collection
func applyMapMethods(filter *Filter) int {
	if filter.item == nil {
		return helpers.ErrorInvalidMethodParameters
	}
	method := filter.methods[0]
	dbEntryData := filter.innerData[len(filter.innerData)-1].(map[string]interface{})
	//
	if item, ok := filter.item.([]interface{}); ok {
		if filter.get {
			switch method {
			case MethodLength:
				filter.methods = filter.methods[1:]
				if len(filter.methods) > 0 {
					// run methods as float64
					filter.innerData = append(filter.innerData, float64(len(dbEntryData)))
					if err := applyNumberMethods(filter); err != 0 {
						return err
					}
					filter.innerData = filter.innerData[:len(filter.innerData) - 1]
				} else {
					filter.item = len(dbEntryData)
				}
				return 0

			case MethodKeyOf:
				if len(item) == 0 {
					return helpers.ErrorInvalidMethodParameters
				}
				var keyOf string
				var err int
				if keyOf, err = mapKeyOf(filter, item[0], dbEntryData); err != 0 {
					return err
				}
				if len(filter.methods) > 0 {
					// run methods as string
					filter.innerData = append(filter.innerData, keyOf)
					if err := applyStringMethods(filter); err != 0 {
						return err
					}
					filter.innerData = filter.innerData[:len(filter.innerData) - 1]
				} else {
					filter.item = keyOf
				}
				return 0

			case MethodContains:
				if len(item) == 0 {
					return helpers.ErrorInvalidMethodParameters
				}
				var keyOf string
				var err int
				if keyOf, err = mapKeyOf(filter, item[0], dbEntryData); err != 0 {
					return err
				}
				filter.methods = []string{}
				filter.item = (keyOf != "")
				return 0
			}
		} else {
			// Update methods
			switch method {
			case MethodDelete:
				// Delete parameters - eg: ["Mary", "Joe", "Vokome"]
				if mParams, ok := item[0].([]interface{}); ok {
					for _, n := range mParams {
						if itemName, ok := n.(string); ok {
							delete(dbEntryData, itemName)
						} else {
							return helpers.ErrorInvalidMethodParameters
						}
					}
					filter.methods = filter.methods[1:]
					if len(filter.methods) > 0 {
						filter.item = item[1:]
						filter.innerData[len(filter.innerData) - 1] = dbEntryData
						if err := applyMapMethods(filter); err != 0 {
							return err
						}
						return 0
					}
					filter.item = dbEntryData
					return 0
				}
				return helpers.ErrorInvalidMethodParameters

			case MethodAppend:
				// Append method - eg: {"x": 27, "y": 43}
				if len(item) == 0 {
					return helpers.ErrorNotEnoughMethodParameters
				}
				if mParams, ok := item[0].(map[string]interface{}); ok {
					// Disallow methods on append items
					m := append([]string{}, filter.methods[1:]...)
					filter.methods = []string{}
					filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem).dataType)
					var itemName string
					for itemName, filter.item = range mParams {
						if iTypeErr := queryItemFilter(filter); iTypeErr != 0 {
							return iTypeErr
						}
						dbEntryData[itemName] = filter.item
						filter.innerData[len(filter.innerData)-1].(map[string]interface{})[itemName] = filter.item
					}
					filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
					if len(m) > 0 {
						filter.methods = m
						filter.item = item[1:]
						filter.innerData[len(filter.innerData) - 1] = dbEntryData
						if err := applyMapMethods(filter); err != 0 {
							return err
						}
						return 0
					}
					filter.item = dbEntryData
					return 0
				}
				return helpers.ErrorInvalidMethodParameters
			}
		}
	}

	// Checking for item with the name method[0] (Items with * not accepted)
	if !strings.Contains(method, "*") {
		filter.methods = filter.methods[1:]
		filter.schemaItems = append(filter.schemaItems, filter.schemaItems[len(filter.schemaItems)-1].iType.(MapItem).dataType)
		filter.innerData = append(filter.innerData, dbEntryData[method])
		if iTypeErr := queryItemFilter(filter); iTypeErr != 0 {
			return iTypeErr
		}
		filter.innerData = filter.innerData[:len(filter.innerData)-1]
		filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
		if !filter.get {
			dbEntryData[method] = filter.item
			filter.item = dbEntryData
			return 0
		}
		return 0
	}
	return helpers.ErrorInvalidMethod
}

func mapKeyOf(filter *Filter, searchItem interface{}, dbEntryData map[string]interface{}) (string, int) {
	// Get inner data type
	si := filter.schemaItems[len(filter.schemaItems) - 1].iType.(MapItem).dataType
	var keyOf string
	if si.IsNumeric() {
		var ok bool
		if searchItem, ok = makeType(searchItem, si); !ok {
			return "", helpers.ErrorInvalidMethodParameters
		}
		for key, innerItem := range dbEntryData {
			if innerItem, ok = makeType(innerItem, si); !ok {
				return "", helpers.ErrorUnexpected
			}
			if searchItem == innerItem {
				keyOf = key
				break
			}
		}
	} else if si.typeName == ItemTypeString || si.typeName == ItemTypeBool {
		for key, innerItem := range dbEntryData {
			if searchItem == innerItem {
				keyOf = key
				break
			}
		}
	} else {
		return "", helpers.ErrorInvalidMethod
	}
	filter.item = filter.item.([]interface{})[1:]
	filter.methods = filter.methods[1:]
	return keyOf, 0
}

// Run methods on Object item
func applyObjectMethods(filter *Filter) int {
	method := filter.methods[0]
	// Check if valid object item
	si := (filter.schemaItems[len(filter.schemaItems)-1].iType.(ObjectItem).schema)[method]
	if !si.Validate() {
		return helpers.ErrorInvalidMethod
	}
	// Remove this method and add new schemaItem
	filter.methods = filter.methods[1:]
	dbEntryData := filter.innerData[len(filter.innerData)-1].([]interface{})
	filter.schemaItems = append(filter.schemaItems, si)
	filter.innerData = append(filter.innerData, dbEntryData[si.dataIndex])
	iTypeErr := queryItemFilter(filter)
	if iTypeErr != 0 {
		return iTypeErr
	}
	filter.innerData = filter.innerData[:len(filter.innerData)-1]
	filter.schemaItems = filter.schemaItems[:len(filter.schemaItems) - 1]
	if !filter.get {
		dbEntryData[si.dataIndex] = filter.item
		filter.item = dbEntryData
	}
	return 0
}

// Run methods on Time item
func applyTimeMethods(filter *Filter, t time.Time) int {
	// Get method duration
	var d time.Duration
	switch filter.methods[0] {
	case MethodSince:
		d = time.Since(t)

	case MethodUntil:
		d = time.Until(t)

	default:
		return helpers.ErrorInvalidMethod
	}

	// Get method format
	format := MethodSecond
	if len(filter.methods) > 1 {
		format = filter.methods[1]
	}

	switch format {
	case MethodMillisecond:
		filter.item = d.Seconds() * 1000

	case MethodSecond:
		filter.item = d.Seconds()

	case MethodMinute:
		filter.item = d.Minutes()

	case MethodHour:
		filter.item = d.Hours()

	case MethodDay:
		filter.item = d.Hours() / 24

	default:
		return helpers.ErrorInvalidMethod
	}

	//
	filter.methods = []string{}
	return 0
}