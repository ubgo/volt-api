// Code generated from Pkl module `Database`. DO NOT EDIT.
package databaseconfig

type Setting struct {
	// The username for this database.
	Username string `pkl:"username"`

	// The password for this database.
	Password string `pkl:"password"`

	// The remote host for this database.
	Host string `pkl:"host"`

	// The remote port for this database.
	Port uint16 `pkl:"port"`

	// The name of the database.
	DbName string `pkl:"dbName"`

	Sslmode string `pkl:"sslmode"`
}
