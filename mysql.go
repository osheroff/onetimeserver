package onetimeserver

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

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

func (m *Mysql) setupMysqlPath() (string, error) {
	tmpPath := fmt.Sprintf("%s/.onetimeserver/tmp", os.Getenv("HOME"))

	if err := os.MkdirAll(tmpPath, 0755); err != nil {
		return "", err
	}

	path, tErr := ioutil.TempDir(tmpPath, "mysql")
	abortOnError(tErr)

	return path, nil
}

func (m *Mysql) mysqlInstallDB() {
	GetBinary("mysql", "/bin", "my_print_defaults", "5.6.26")
	GetBinary("mysql", "/bin", "resolveip", "5.6.26")
	GetBinary("mysql", "/support-files", "my-default.cnf", "5.6.26")

	sqlFiles := [...]string{"fill_help_tables.sql", "mysql_security_commands.sql", "mysql_system_tables.sql", "mysql_system_tables_data.sql", "errmsg.sys"}
	for _, sql := range sqlFiles {
		GetBinary("mysql", "/share", sql, "5.6.26")
	}

	binPath := GetBinary("mysql", "", "mysql_install_db", "5.6.26")

	cmd := exec.Command(binPath,
		fmt.Sprintf("--datadir=%s", m.path),
		fmt.Sprintf("--basedir=%s", filepath.Dir(binPath)),
		"--user=ben",
		"--no-defaults")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	abortOnError(cmd.Run())
}

func (m *Mysql) Boot(args []string) (map[string]interface{}, error) {
	var err error
	infoMap := make(map[string]interface{})

	execPath := GetBinary("mysql", "/bin", "mysqld", "5.6.26")

	m.port = GetPort(33306)

	m.path, err = m.setupMysqlPath()
	if err != nil {
		return infoMap, err
	}

	m.mysqlInstallDB()

	infoMap["mysql_path"] = m.path
	infoMap["port"] = m.port

	defaultArgs := []string{
		fmt.Sprintf("--lc-messages-dir=%s", filepath.Dir(GetBinary("mysql", "/share", "errmsg.sys", "5.6.26"))),
		fmt.Sprintf("--datadir=%s", m.path),
		fmt.Sprintf("--port=%d", m.port)}

	newArgs := append(defaultArgs, args...)
	m.cmd = exec.Command(execPath, newArgs...)

	stderr, err := m.cmd.StderrPipe()
	m.cmd.Stdout = os.Stdout

	err = m.cmd.Start()
	abortOnError(err)

	scanner := bufio.NewScanner(stderr)

	go m.cmd.Process.Wait()

	for scanner.Scan() {
		if matched, _ := regexp.Match(`^Version:\s+'[\d\.]+'.*`, scanner.Bytes()); matched {
			break
		}
		fmt.Printf("[mysqld] %s\n", scanner.Text())
	}

	m.pid = m.cmd.Process.Pid
	return infoMap, nil
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
