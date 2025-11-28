package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"golang.org/x/mod/modfile"

	"go.mau.fi/util/exerrors"
)

const ldflagTemplate = "-s -w -X '%[1]s.Tag=%[2]s' -X '%[1]s.Commit=%[3]s' -X '%[1]s.BuildTime=%[4]s' -X 'maunium.net/go/mautrix.GoModVersion=%[5]s' %[6]s"

func main() {
	versionPackage := "main"
	var gitCommit, gitTag string
	if os.Getenv("CI") == "true" {
		gitCommit = os.Getenv("CI_COMMIT_SHA")
		gitTag = os.Getenv("CI_COMMIT_TAG")
	} else {
		gitCommit = subcommand("git", "rev-parse", "HEAD")
		gitTag = subcommand("git", "describe", "--exact-match", "--tags")
	}
	extraLDFlags := os.Getenv("GO_LDFLAGS")
	if os.Getenv("MAU_STATIC_BUILD") == "true" {
		extraLDFlags += " -linkmode external -extldflags '-static'"
	}
	ldflags := fmt.Sprintf(
		ldflagTemplate,
		versionPackage,
		gitTag,
		gitCommit,
		time.Now().Format(time.RFC3339),
		getMautrixGoVersion(),
		extraLDFlags,
	)
	args := []string{"go", "build", "-ldflags", ldflags}
	args = append(args, os.Args[1:]...)
	buildPackage := os.Getenv("MAU_BUILD_PACKAGE_OVERRIDE")
	if buildPackage == "" {
		buildPackage = "./cmd/" + os.Getenv("BINARY_NAME")
	}
	args = append(args, buildPackage)
	env := os.Environ()
	if runtime.GOOS == "darwin" && os.Getenv("LIBRARY_PATH") == "" {
		if brewPrefix := subcommand("brew", "--prefix"); brewPrefix != "" {
			fmt.Println("Mac: Using", brewPrefix, "for LIBRARY_PATH and CPATH")
			env = append(env, fmt.Sprintf("LIBRARY_PATH=%s/lib", brewPrefix))
			env = append(env, fmt.Sprintf("CPATH=%s/include", brewPrefix))
		} else if directoryExists("/opt/homebrew") {
			fmt.Println("Mac: Using /opt/homebrew for LIBRARY_PATH and CPATH")
			env = append(env, "LIBRARY_PATH=/opt/homebrew/lib")
			env = append(env, "CPATH=/opt/homebrew/include")
		}
	} else if strings.HasSuffix(runtime.GOOS, "bsd") && os.Getenv("LIBRARY_PATH") == "" {
		fmt.Println("BSD: Using /usr/local for LIBRARY_PATH and CPATH")
		env = append(env, "LIBRARY_PATH=/usr/local/bin")
		env = append(env, "CPATH=/usr/local/include")
	}

	goBin := exerrors.Must(exec.LookPath("go"))
	fmt.Println("Running", string(exerrors.Must(json.Marshal(args))))
	exerrors.PanicIfNotNil(syscall.Exec(goBin, args, env))
}

func directoryExists(dir string) bool {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func getMautrixGoVersion() string {
	parsedGoMod := exerrors.Must(modfile.Parse("go.mod", exerrors.Must(os.ReadFile("go.mod")), nil))
	for _, req := range parsedGoMod.Require {
		if req.Mod.Path == "maunium.net/go/mautrix" {
			return req.Mod.Version
		}
	}
	return ""
}

func subcommand(command string, args ...string) string {
	stdout, _ := exec.Command(command, args...).Output()
	return strings.TrimSpace(string(stdout))
}
