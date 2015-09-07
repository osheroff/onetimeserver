package onetimeserver

import "fmt"

type Mysql struct {
	port int
	path string
	pid  int
}

func NewMysql() *Mysql {
	return &Mysql{}
}

func (m *Mysql) Boot(args []string) {
	execPath := GetBinary("mysql", "mysqld", "5.6.26")
	fmt.Printf(execPath)
	m.port = GetPort(33306)
	m.pid = 7209
}

func (m *Mysql) Kill() {
	fmt.Printf("IMMA KILLIN %d\n", m.pid)
}

func (m *Mysql) String() string {
	return fmt.Sprintf("mysql {port: %d, pid: %d}", m.port, m.pid)
}

func (m *Mysql) Pid() int {
	return m.pid
}

func (m *Mysql) Port() int {
	return m.port
}
