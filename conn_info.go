package connassist

import (
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
)

// GetMySQLConnInfoIgnoreErr 获取MySQL连接的字符串显示，忽略错误信息
func GetMySQLConnInfoIgnoreErr(conn *sql.DB) (connInfo string) {
	connInfo, _ = GetMySQLConnInfo(conn)
	return
}

// GetMySQLConnInfo 获取MySQL连接的字符串显示
func GetMySQLConnInfo(conn *sql.DB) (connInfo string, err error) {
	defer func() {
		if panicRecover := recover(); panicRecover != nil {
			err = fmt.Errorf(
				"get mysql connection info failed for %v", panicRecover)
			debug.PrintStack()
		}
	}()

	cv := reflect.ValueOf(conn).Elem()
	// fmt.Printf("%#v\n", cv)

	dsnStr := getMySQLConnDSN(conn)
	dsnInfo, err := resolveDsn(dsnStr)
	if err != nil {
		return
	}

	fcsv := cv.FieldByName("dep")
	// fmt.Printf("%#v\n", fcsv)

	lports := make([]string, 0)
	for _, key := range fcsv.MapKeys() {
		// fcv := fcsv.Index(0).Elem()
		depSet := fcsv.MapIndex(key).MapKeys()
		if len(depSet) < 1 {
			continue
		}

		fcv := depSet[0].Elem().Elem()
		// fmt.Printf("%#v\n", fcv)
		// fmt.Printf("%#v\n", fcv.Elem())

		mc := fcv.FieldByName("ci").Elem().Elem()
		// fmt.Printf("%#v", mc)

		nc := mc.FieldByName("netConn").Elem().Elem()
		// fmt.Printf("%#v", nc)

		ic := nc.FieldByName("conn")
		// fmt.Printf("%#v", conn)

		fd := ic.FieldByName("fd").Elem()
		// fmt.Printf("%#v", fd)

		la := fd.FieldByName("laddr").Elem().Elem()
		// fmt.Printf("%#v", la)

		lport := la.FieldByName("Port")

		lports = append(lports, fmt.Sprint(lport))
		// fmt.Printf("%#v", lport)
	}

	connInfo = fmt.Sprintf(
		"127.0.0.1:%s <==> %s:%s", strings.Join(lports, ", "),
		dsnInfo["host"], dsnInfo["port"])
	return
}

func resolveDsn(dsn string) (info map[string]string, err error) {
	info = make(map[string]string)

	// r, _ := regexp.Compile("p(?P<haha>[a-z]+)ch")
	r, err := regexp.Compile(`.*:.*@tcp\((?P<host>[\w.]+):(?P<port>\d+)\)/`)
	if err != nil {
		return
	}

	subStrs := r.FindStringSubmatch(dsn)
	for idx, name := range r.SubexpNames() {
		if len(name) < 1 {
			continue
		}

		info[name] = subStrs[idx]
	}

	return
}
