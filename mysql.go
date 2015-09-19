package onetimeserver

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type Mysql struct {
	port    int
	path    string
	pid     int
	version string
	cmd     *exec.Cmd
}

func NewMysql(version string) *Mysql {
	mappedVersion, err := mapVersion(version)
	abortOnError(err)

	return &Mysql{version: mappedVersion}
}

func mapVersion(version string) (string, error) {
	availableVersions := map[string]string{
		"5.6":    "5.6.26",
		"5.6.26": "5.6.26",
		"5.5":    "5.5.45",
		"5.5.45": "5.5.45",
	}

	if version == "" {
		return "5.6.26", nil
	}

	if availableVersions[version] == "" {
		return "", errors.New(fmt.Sprintf("no such mysql version: %s", version))
	}

	return availableVersions[version], nil
}

func (m *Mysql) getMysqlBinary(path string, bin string) string {
	return GetBinary("mysql", path, bin, m.version)
}

func abortOnError(e error) {
	if e != nil {
		log.Fatal("_onetimeserver_json: { \"error\": \"%s\" }\n", e)
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
	m.getMysqlBinary("/bin", "my_print_defaults")
	m.getMysqlBinary("/bin", "resolveip")
	if m.version > "5.5.45" {
		m.getMysqlBinary("/share", "mysql_security_commands.sql")
		m.getMysqlBinary("/support-files", "my-default.cnf")
	}

	m.getMysqlBinary("/share", "errmsg.sys")
	m.getMysqlBinary("/share/english", "errmsg.sys")

	sqlFiles := [...]string{"fill_help_tables.sql", "mysql_system_tables.sql", "mysql_system_tables_data.sql", "errmsg.sys"}
	for _, sql := range sqlFiles {
		m.getMysqlBinary("/share", sql)
	}

	binPath := m.getMysqlBinary("", "mysql_install_db")

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

	execPath := m.getMysqlBinary("/bin", "mysqld")

	m.port = GetPort(33306)

	m.path, err = m.setupMysqlPath()
	if err != nil {
		return infoMap, err
	}

	m.mysqlInstallDB()

	infoMap["mysql_path"] = m.path
	infoMap["port"] = m.port

	defaultArgs := []string{
		fmt.Sprintf("--lc-messages-dir=%s", filepath.Dir(m.getMysqlBinary("/share/english", "errmsg.sys"))),
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
	process, err := os.FindProcess(m.pid)
	if err == nil {
		fmt.Printf("killing %d\n", m.pid)
		process.Kill()
	} else {
		fmt.Printf("Couldn't find process %d -- %s\n", m.pid, err)
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
