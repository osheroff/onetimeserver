package onetimeserver

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Mysql struct {
	port    int
	path    string
	pid     int
	version string
	reuse   string
	debug   bool
	cmd     *exec.Cmd
}

func NewMysql(version string, reuse string, debug bool) *Mysql {
	mappedVersion, err := mapVersion(version)
	abortOnError(err)

	return &Mysql{version: mappedVersion, reuse: reuse, debug: debug}
}

func mapVersion(version string) (string, error) {
	availableVersions := map[string]string{
		"5.6":    "5.6.26",
		"5.6.26": "5.6.26",
		"5.7":    "5.7.17",
		"5.7.17": "5.7.17",
		"5.5":    "5.5.45",
		"5.5.45": "5.5.45",
		"8.0":    "8.0.12",
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
		log.Fatalf("_onetimeserver_json: { \"error\": \"%s\" }\n", e)
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

func (m *Mysql) pullBinaries() {
	m.getMysqlBinary("/bin", "resolveip")

	if m.version >= "8.0" && runtime.GOOS == "darwin" {
		m.getMysqlBinary("/lib", "libssl.1.0.0.dylib")
		m.getMysqlBinary("/lib", "libcrypto.1.0.0.dylib")
	}

	if runtime.GOOS == "linux" {
		m.getMysqlBinary("/bin", "libaio.so.1")
	}

	if m.version > "5.5.45" && m.version < "8.0" {
		m.getMysqlBinary("/support-files", "my-default.cnf")
	}

	if runtime.GOOS == "linux" {
		m.getMysqlBinary("/bin", "libaio.so.1")
	}

	m.getMysqlBinary("/share", "errmsg.sys")
	m.getMysqlBinary("/share/english", "errmsg.sys")

	sqlFiles := [...]string{"errmsg.sys"}
	for _, sql := range sqlFiles {
		m.getMysqlBinary("/share", sql)
	}
}

// useful for generating the tarball
func (m *Mysql) _mysqlInstallDB() {
	m.getMysqlBinary("/bin", "my_print_defaults")

	if m.version > "5.5.45" && m.version < "8.0" {
		m.getMysqlBinary("/share", "mysql_security_commands.sql")
	}

	sqlFiles := [...]string{"fill_help_tables.sql", "mysql_system_tables.sql", "mysql_system_tables_data.sql", "errmsg.sys"}
	for _, sql := range sqlFiles {
		m.getMysqlBinary("/share", sql)
	}

	var cmd *exec.Cmd
	var binPath string
	if m.version >= "5.7" {
		binPath = m.getMysqlBinary("/bin", "mysqld")
		cmd = exec.Command(binPath,
			"--initialize-insecure",
			fmt.Sprintf("--datadir=%s", m.path),
			fmt.Sprintf("--basedir=%s", filepath.Dir(binPath)))
	} else {
		binPath = m.getMysqlBinary("", "mysql_install_db")
		cmd = exec.Command(binPath,
			fmt.Sprintf("--datadir=%s", m.path),
			fmt.Sprintf("--basedir=%s", filepath.Dir(binPath)))
		// "--no-defaults")

		fmt.Println(m.getMysqlBinary("/bin", "my_print_defaults"))
		str := fmt.Sprintf("LD_LIBRARY_PATH=%s", filepath.Dir(m.getMysqlBinary("/bin", "my_print_defaults")))
		fmt.Println(str)
		cmd.Env = append(cmd.Env, str)
	}

	cmd.Dir = filepath.Dir(binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if m.debug {
		fmt.Printf("executing %v\n", cmd)
	}

	abortOnError(cmd.Run())
}

func (m *Mysql) mysqlInstallDB() {
	m.pullBinaries()
	tarballPath := m.getMysqlBinary("", "installed_db.tar.gz")

	cmd := exec.Command("tar", "zxf", tarballPath, "-C", m.path, "--strip-components=1")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	abortOnError(cmd.Run())
}

func (m *Mysql) Boot(args []string) (map[string]interface{}, error) {
	var err error
	infoMap := make(map[string]interface{})
	log.Printf("booting mysql server (version %s)", m.version)

	execPath := m.getMysqlBinary("/bin", "mysqld")

	hasServerID := false

	for _, arg := range args {
		if strings.HasPrefix(arg, "--server-id") || strings.HasPrefix(arg, "--server_id") {
			hasServerID = true
		} else if strings.HasPrefix(arg, "--port=") {
			i, e := strconv.Atoi(arg[len("--port="):len(arg)])
			if e == nil {
				m.port = i
			}
		}
	}

	if m.port == 0 {
		m.port = TryPort(33306)
		if m.port == 0 {
			m.port = GetPort()
		}
	}

	if m.reuse == "" {
		m.path, err = m.setupMysqlPath()
		if err != nil {
			return infoMap, err
		}

		m.mysqlInstallDB()
	} else {
		m.path = m.reuse
	}

	infoMap["mysql_path"] = m.path
	infoMap["port"] = m.port

	defaultArgs := []string{
		"--no-defaults",
		"--bind-address=127.0.0.1",
		"--innodb-buffer-pool-size=10M",
		"--performance_schema=0",
		"--innodb_use_native_aio=0",
		fmt.Sprintf("--lc-messages-dir=%s", filepath.Dir(m.getMysqlBinary("/share/english", "errmsg.sys"))),
		fmt.Sprintf("--socket=%s/mysql.sock", m.path),
		fmt.Sprintf("--datadir=%s", m.path),
		fmt.Sprintf("--port=%d", m.port)}

	newArgs := append(defaultArgs, args...)

	if !hasServerID {
		newArgs = append(newArgs, fmt.Sprintf("--server_id=%d", rand.Int31()))
	}

	m.cmd = exec.Command(execPath, newArgs...)

	stderr, err := m.cmd.StderrPipe()
	m.cmd.Stdout = os.Stdout

	// add path for lbiaio.so
	m.cmd.Env = []string{fmt.Sprintf("LD_LIBRARY_PATH=%s", filepath.Dir(execPath))}

	err = m.cmd.Start()
	abortOnError(err)

	scanner := bufio.NewScanner(stderr)

	go m.cmd.Process.Wait()

	for scanner.Scan() {
		if matched, _ := regexp.Match(`Version:\s+'[\d\.]+.*`, scanner.Bytes()); matched {
			break
		}
		fmt.Printf("[mysqld] %s\n", scanner.Text())
	}

	m.pid = m.cmd.Process.Pid
	return infoMap, nil
}

func (m *Mysql) Kill(cleanup bool) {
	process, err := os.FindProcess(m.pid)
	if err == nil {
		fmt.Printf("killing %d\n", m.pid)
		process.Signal(syscall.SIGTERM)
		time.Sleep(3 * time.Second)
		process.Kill()
	} else {
		fmt.Printf("Couldn't find process %d -- %s\n", m.pid, err)
	}

	if m.path != "" && cleanup {
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
