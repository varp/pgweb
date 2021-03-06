package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
)

var (
	testCommands   map[string]string
	ServerHost     string
	ServerPort     string
	ServerUser     string
	ServerPassword string
	ServerDatabase string
)

func pgVersion() (int, int) {
	var major, minor int
	fmt.Sscanf(os.Getenv("PGVERSION"), "%d.%d", &major, &minor)
	return major, minor
}

func getVar(name, def string) string {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	return val
}

func initVars() {
	ServerHost = getVar("PGHOST", "localhost")
	ServerPort = getVar("PGPORT", "15432")
	ServerUser = getVar("PGUSER", "postgres")
	ServerPassword = getVar("PGPASSWORD", "postgres")
	ServerDatabase = getVar("PGDATABASE", "booktown")
}

func setupCommands() {
	testCommands = map[string]string{
		"createdb": "createdb",
		"psql":     "psql",
		"dropdb":   "dropdb",
	}

	if onWindows() {
		for k, v := range testCommands {
			testCommands[k] = v + ".exe"
		}
	}
}

func onWindows() bool {
	return runtime.GOOS == "windows"
}

func setup() {
	out, err := createDb()

	if err != nil {
		fmt.Println(fmt.Errorf("database creation failed: %s, err: %e", string(out), err))

		var (
			o []byte // out
			e error  // error
		)

		killConnections()

		o, e = dropDb()
		if e != nil {
			fmt.Println(fmt.Errorf("dataase dropping failed: %s, err: %s", string(o), e.Error()))
			return
		}

		o, e = createDb()
		if e != nil {
			fmt.Println(fmt.Errorf("dataase dropping failed: %s, err: %s", string(o), e.Error()))
			return
		}
	}

	_, filename, _, _ := runtime.Caller(1)

	out, err = exec.Command(
		testCommands["psql"],
		"-U", ServerUser,
		"-h", ServerHost,
		"-p", ServerPort,
		"-f", path.Join(path.Dir(filename), "../../data/booktown.sql"),
		ServerDatabase,
	).CombinedOutput()

	if err != nil {
		fmt.Println("Database import failed:", string(out))
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func createDb() ([]byte, error) {
	return exec.Command(
		testCommands["createdb"],
		"-U", ServerUser,
		"-h", ServerHost,
		"-p", ServerPort,
		ServerDatabase,
	).CombinedOutput()
}

func killConnections() {

	killConnectionsCmd := "select pg_terminate_backend(pid) from pg_stat_activity where datname = '%s'"

	_, err := exec.Command(
		testCommands["psql"],
		"-U", "postgres",
		"-h", ServerHost,
		"-p", ServerPort,
		"postgres",
		"-c", fmt.Sprintf(killConnectionsCmd, ServerDatabase),
	).CombinedOutput()

	if err != nil {
		fmt.Println("Teardown error:", err)
	}
}

func teardown() {

	killConnections()

	_, err := dropDb()

	if err != nil {
		fmt.Println("Teardown error:", err)
	}
}

func dropDb() ([]byte, error) {
	return exec.Command(
		testCommands["dropdb"],
		"-U", ServerUser,
		"-h", ServerHost,
		"-p", ServerPort,
		ServerDatabase,
	).CombinedOutput()
}

func CreateBooktownDB() {
	initVars()
	setupCommands()

	setup()
}

func DropBooktownDb() {
	teardown()
}

func init() {
	initVars()
}
