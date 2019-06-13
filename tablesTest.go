package main

import (
	"github.com/hewiefreeman/GopherGameDB/schema"
	"github.com/hewiefreeman/GopherGameDB/userTable"
	//"github.com/hewiefreeman/GopherGameDB/helpers"
	"encoding/json"
	"strconv"
	"fmt"
	"time"
)

func main() {
	// JSON query and unmarshalling test
	newTableJson := "{\"NewUserTable\": [\"users\",{\"email\": [\"String\", \"\", 0, true, true],\"friends\": [\"Array\", [\"Object\", {\"name\": [\"String\", \"\", 0, true, true],\"status\": [\"Uint8\", 0, 0, 2, false]}, false], 50, false],\"vCode\": [\"String\", \"\", 0, true, false],\"verified\": [\"Bool\", false], \"mmr\": [\"Uint16\", 1500, 1100, 2250, false], \"testMap\": [\"Map\", [\"Map\", [\"Uint16\", 100, 0, 0, false], 0, false], 0, false], \"timeStamp\": [\"Time\", \"Kitchen\", false]}, 0, 0, 0, 0]}"
	v := make(map[string]interface{})
	err := json.Unmarshal([]byte(newTableJson), &v)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get the schema object from the query
	s := v["NewUserTable"].([]interface{})[1].(map[string]interface{})

	// Make a schema with the query's schema object
	schemaObj, schemaErr := schema.New(s)
	if schemaErr != 0 {
		fmt.Println("Schema Error:", schemaErr)
		return
	}

	// Make a new UserTable with the schema
	table, tableErr := userTable.New("users", schemaObj, 6000, 0, 0, 0, 0)
	if tableErr != 0 {
		fmt.Println("Table Create Error:", tableErr)
		return
	}

	// insert
	var averageTime float64
	for v := 0; v < 100; v++ {
		now := time.Now()
		// Insert a test User
		insertErr := table.NewUser("guest"+strconv.Itoa(v), "myPass", map[string]interface{}{"email": "dinospumoni"+strconv.Itoa(v)+"@yahoo.com", "mmr": 1674, "vCode": "06AJ3T9"})
		if insertErr != 0 {
			fmt.Println("Insert Error:", insertErr)
			return
		}
		if averageTime == 0 {
			averageTime = time.Since(now).Seconds()
		} else {
			averageTime = (averageTime + time.Since(now).Seconds()) / 2
		}
	}
	fmt.Println("Average insert time (ms):", averageTime*1000)

	averageTime = 0
	for v := 0; v < 100; v++ {
		now := time.Now()
		// add 1 to entry's mmr
		updateErr := table.UpdateUserData("guest"+strconv.Itoa(v), "myPass", map[string]interface{}{"mmr.*add": []interface{}{2}})
		if updateErr != 0 {
			fmt.Println("Update Error:", updateErr)
			return
		}
		if averageTime == 0 {
			averageTime = time.Since(now).Seconds()
		} else {
			averageTime = (averageTime + time.Since(now).Seconds()) / 2
		}
	}
	fmt.Println("Average update time (ms):", averageTime*1000)

	averageTime = 0
	for v := 0; v < 100; v++ {
		now := time.Now()
		_, ueErr := table.GetUserData("guest"+strconv.Itoa(v), "myPass")
		if ueErr != 0 {
			fmt.Println("User Data Error:", ueErr)
			return
		}
		if averageTime == 0 {
			averageTime = time.Since(now).Seconds()
		} else {
			averageTime = (averageTime + time.Since(now).Seconds()) / 2
		}
	}
	fmt.Println("Average select time (ms):", averageTime*1000)

	ud, ueErr := table.GetUserData("guest99", "myPass")
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
	fmt.Println("Before:", ud)

	// Multiply by 1.5, divide by 2, add 4, then subtract 1 from entry's mmr (using methods)
	updateErr := table.UpdateUserData("guest99", "myPass", map[string]interface{}{"mmr.*mul.*div.*add.*sub": []interface{}{1.5, 2, 4, 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 1:", updateErr)
		return
	}

	// Append a friend to friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*append": []interface{}{map[string]interface{}{"name": "Mag"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 2:", updateErr)
		return
	}

	// Prepend a friend to friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*prepend": []interface{}{map[string]interface{}{"name": "Jason"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 3:", updateErr)
		return
	}

	// Append 2 friends to index 1 of friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*append[1]": []interface{}{map[string]interface{}{"name": "Harry"}, map[string]interface{}{"name": "Potter"}}})
	if updateErr != 0 {
		fmt.Println("Update Error 4:", updateErr)
		return
	}

	// Delete 2 friends from friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.*delete": []interface{}{2, 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 5:", updateErr)
		return
	}

	// Chage name of friend at index 1 of friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.1.name.*append": []interface{}{"icarp"}})
	if updateErr != 0 {
		fmt.Println("Update Error 6:", updateErr)
		return
	}

	// Chage status of friend at index 0 of friends
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"friends.0.status": 3})
	if updateErr != 0 {
		fmt.Println("Update Error 7:", updateErr)
		return
	}

	// Add something to testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.*append": map[string]interface{}{"items": map[string]interface{}{}}})
	if updateErr != 0 {
		fmt.Println("Update Error 8:", updateErr)
		return
	}

	// Add something to items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.arrows": 12})
	if updateErr != 0 {
		fmt.Println("Update Error 8:", updateErr)
		return
	}

	// Add something to items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.beanz": 87})
	if updateErr != 0 {
		fmt.Println("Update Error 9:", updateErr)
		return
	}

	// Apply arithmetic to beanz in items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.beanz.*add.*mul": []interface{}{3, 2}})
	if updateErr != 0 {
		fmt.Println("Update Error 10:", updateErr)
		return
	}

	// Delete arrows in items in testMap
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.*delete": []interface{}{"arrows"}})
	if updateErr != 0 {
		fmt.Println("Update Error 11:", updateErr)
		return
	}

	// Append rupees and silk to items
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"testMap.items.*append": map[string]interface{}{"rupees": 99, "silk": 1}})
	if updateErr != 0 {
		fmt.Println("Update Error 12:", updateErr)
		return
	}

	// Set timeStamp manually
	updateErr = table.UpdateUserData("guest99", "myPass", map[string]interface{}{"timeStamp": "4:29PM"})
	if updateErr != 0 {
		fmt.Println("Update Error 12:", updateErr)
		return
	}

	ud, ueErr = table.GetUserData("guest99", "myPass")
	if ueErr != 0 {
		fmt.Println("User Data Error:", ueErr)
		return
	}
	fmt.Println("After:", ud)
}
