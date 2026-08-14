package crud

import (
	"blueprint/connections"
	"os"

	models "blueprint/models"

	"github.com/aquasecurity/table"
	"github.com/liamg/tml"
)

func getTypeName(dbtype models.DatabaseType) string {
	var Name string = ""
	for i := range models.DatabaseTypes {
		if models.DatabaseTypes[i].Value == dbtype {
			Name = models.DatabaseTypes[i].Name
		}
	}
	return Name
}

func List(input models.CommandInput) {

	t := table.New(os.Stdout)
	t.SetRowLines(false)

	t.SetHeaders(tml.Sprintf("<yellow>CONNECTIONS</yellow>"))
	t.AddHeaders(
		tml.Sprintf("<white>Name</white>"),
		tml.Sprintf("<white>Type</white>"),
		tml.Sprintf("<white>Server</white>"),
		tml.Sprintf("<white>Port</white>"),
		tml.Sprintf("<white>Database</white>"),
		tml.Sprintf("<white>User</white>"))
	//tml.Sprintf("<white>Password</white>"))

	t.SetHeaderColSpans(0, 6)
	t.SetAutoMergeHeaders(true)
	t.SetHeaderStyle(table.StyleBold)
	t.SetLineStyle(table.StyleWhite)

	config, err := connections.LoadConnections()
	if err != nil {
		panic(err)
	}

	for _, conn := range config.Connections {

		t.AddRow(
			tml.Sprintf("<green>"+conn.Name+"</green>"),
			getTypeName(conn.Type),
			conn.Server,
			conn.Port,
			conn.Database,
			conn.User,
		)
	}

	t.Render()
}
