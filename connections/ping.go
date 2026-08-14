package connections

import (
	"blueprint/database"
	"blueprint/models"
	"fmt"
	"strings"
)

func PadRight(text string, width int) string {
	if len(text) >= width {
		return text
	}

	return text + strings.Repeat(" ", width-len(text))
}

func Test(input models.CommandInput) {

	var Argument string = ""
	if len(input.Arguments) > 0 {
		Argument = input.Arguments[0]
	}

	id, conn := GetConnection(Argument)

	if id == "" || conn == nil {
		//fmt.Println("❌ Failed to connect!")
		return
	}

	db, err := database.Connect(*conn)

	if err != nil {
		fmt.Println("❌ Failed to connect!")
		//panic(err)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("❌ Failed to connect!")
		//panic(err)
		return
	}

	err = sqlDB.Ping()
	if err != nil {
		fmt.Println("❌ Failed to connect!")
		//panic(err)
		return
	}

	fmt.Println("✅ Connected!")

}
