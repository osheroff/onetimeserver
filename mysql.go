package onetimeserver

import "fmt"
import "io/ioutil"
import "os/exec"
import "os"

type Mysql struct {
	port int
	path string
	pid  int
	cmd  *exec.Cmd
}

func NewMysql() *Mysql {
	return &Mysql{}
}

func abortOnError(e error) {
	if e != nil {
		fmt.Printf("error: %s\n", e)
		os.Exit(1)
	}
}

func setupMysqlPath() string {
	tmpPath := fmt.Sprintf("%s/.onetimeserver/tmp", os.Getenv("HOME"))

	if err := os.MkdirAll(tmpPath, 0755); err != nil {
		abortOnError(err)
	}

	path, tErr := ioutil.TempDir(tmpPath, "mysql")
	abortOnError(tErr)

	return path
}

func (m *Mysql) Boot(args []string) {
	execPath := GetBinary("mysql", "mysqld", "5.6.26")

	m.port = GetPort(33306)

	m.path = setupMysqlPath()
	fmt.Printf("data_path: %s\n", m.path)

	m.cmd = exec.Command(execPath, args...)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		fmt.Printf("err: %s\n", err)
	}

	go m.cmd.Process.Wait()

	m.pid = m.cmd.Process.Pid
}

func (m *Mysql) Kill() {
	if process, err := os.FindProcess(m.pid); err != nil {
		process.Kill()
	}

	if m.path != "" {
		os.RemoveAll(m.path)
	}
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
