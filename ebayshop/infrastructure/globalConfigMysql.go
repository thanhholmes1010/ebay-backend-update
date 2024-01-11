package infrastructure

import "github.com/go-sql-driver/mysql"

var MysqlConfig *mysql.Config = &mysql.Config{
	Addr:                 "localhost:3306",
	Net:                  "tcp",
	Passwd:               "thaianh1711",
	User:                 "theloop",
	DBName:               "ebayclonedb",
	AllowNativePasswords: true,
	MultiStatements:      true,
}
