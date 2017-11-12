package spec

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)


var (
	testCommands   map[string]string
	serverHost     string
	serverPort     string
	serverUser     string
	serverPassword string
	serverDatabase string
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
	serverHost = getVar("PGHOST", "localhost")
	serverPort = getVar("PGPORT", "5432")
	serverUser = getVar("PGUSER", "postgres")
	serverPassword = getVar("PGPASSWORD", "postgres")
	serverDatabase = getVar("PGDATABASE", "booktown")
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

	initVars()
	setupCommands()

	out, err := exec.Command(
		testCommands["createdb"],
		"-U", serverUser,
		"-h", serverHost,
		"-p", serverPort,
		serverDatabase,
	).CombinedOutput()

	if err != nil {
		fmt.Println("Database creation failed:", string(out))
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	out, err = exec.Command(
		testCommands["psql"],
		"-U", serverUser,
		"-h", serverHost,
		"-p", serverPort,
		"-f", "../../data/booktown.sql",
		serverDatabase,
	).CombinedOutput()

	if err != nil {
		fmt.Println("Database import failed:", string(out))
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}


func teardown() {
	_, err := exec.Command(
		testCommands["dropdb"],
		"-U", serverUser,
		"-h", serverHost,
		"-p", serverPort,
		serverDatabase,
	).CombinedOutput()

	if err != nil {
		fmt.Println("Teardown error:", err)
	}
}

func TestPgweb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pgweb Suite")
}

var agoutiDriver *agouti.WebDriver

var _ = BeforeSuite(func() {
	setup()

	agoutiDriver = agouti.ChromeDriver()

	// https://github.com/onsi/ginkgo/issues/285#issuecomment-290575636

	// The Ginkgo test runner takes over os.Args and fills it with
	// its own flags.  This makes the cobra command arg parsing
	// fail because of unexpected options.  Work around this.

	// Save a copy of os.Args
	//origArgs := os.Args[:]

	// Trim os.Args to only the first arg
	//os.Args = os.Args[:1] // trim to only the first arg, which is the command itself

	// Run the command which parses os.Args
	//pwcli.Run()

	// Restore os.Args
	//os.Args = origArgs[:]

	Expect(agoutiDriver.Start()).To(Succeed())
})

var _ = AfterSuite(func() {
	teardown()

	Expect(agoutiDriver.Stop()).To(Succeed())
})
