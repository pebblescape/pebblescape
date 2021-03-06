// Package gitreceive handles 'smart' Git HTTP requests for Pebblescape in a net/http compatible way.
// Derived from https://gitlab.com/gitlab-org/gitlab-git-http-server.
package gitreceive

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/pebblescape/pebblescape/pkg/paths"
	"github.com/pebblescape/pebblescape/pkg/utils"
)

// GitHandler holds the configuration for a Git HTTP handler.
type GitHandler struct {
	repoRoot  string
	cachePath string
}

type gitService struct {
	method     string
	suffix     string
	handleFunc func(gitEnv, string, string, http.ResponseWriter, *http.Request)
	rpc        string
}

type gitEnv struct {
	App string
}

// GitServices speficies all paths this handler processes.
var GitServices = [...]gitService{
	{"GET", "/info/refs", handleGetInfoRefs, ""},
	{"POST", "/git-upload-pack", handlePostRPC, "git-upload-pack"},
	{"POST", "/git-receive-pack", handlePostRPC, "git-receive-pack"},
}

// NewGitHandler returns a new net/http compatible handler.
func NewGitHandler(repoRoot, cachePath string) *GitHandler {
	return &GitHandler{repoRoot, cachePath}
}

func (h *GitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var g gitService

	// Look for a matching Git service
	foundService := false
	for _, g = range GitServices {
		if r.Method == g.method && strings.HasSuffix(r.URL.Path, g.suffix) {
			foundService = true
			break
		}
	}

	name := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, g.suffix), "/git/")
	if !foundService || !utils.AppNamePattern.MatchString(name) {
		// The protocol spec in git/Documentation/technical/http-protocol.txt
		// says we must return 403 if no matching service is found.
		http.Error(w, "Forbidden", 403)
		return
	}

	repoPath, err := prepareRepo(h.repoRoot, name, h.cachePath)
	if err != nil {

		http.Error(w, err.Error(), 404)
		return
	}

	g.handleFunc(gitEnv{App: name}, g.rpc, repoPath, w, r)
}

func handleGetInfoRefs(env gitEnv, _ string, path string, w http.ResponseWriter, r *http.Request) {
	rpc := r.URL.Query().Get("service")
	if !(rpc == "git-upload-pack" || rpc == "git-receive-pack") {
		// The 'dumb' Git HTTP protocol is not supported
		http.Error(w, "Not Found", 404)
		return
	}

	// Prepare our Git subprocess
	cmd, pipe := gitCommand(env, "git", subCommand(rpc), "--stateless-rpc", "--advertise-refs", path)
	if err := cmd.Start(); err != nil {
		fail500(w, "handleGetInfoRefs", err)
		return
	}
	defer cleanUpProcessGroup(cmd) // Ensure brute force subprocess clean-up

	// Start writing the response
	w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", rpc))
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(200) // Don't bother with HTTP 500 from this point on, just return
	if err := pktLine(w, fmt.Sprintf("# service=%s\n", rpc)); err != nil {
		logError("handleGetInfoRefs response", err)
		return
	}
	if err := pktFlush(w); err != nil {
		logError("handleGetInfoRefs response", err)
		return
	}
	if _, err := io.Copy(w, pipe); err != nil {
		logError("handleGetInfoRefs read from subprocess", err)
		return
	}
	if err := cmd.Wait(); err != nil {
		logError("handleGetInfoRefs wait for subprocess", err)
		return
	}
}

func handlePostRPC(env gitEnv, rpc string, path string, w http.ResponseWriter, r *http.Request) {
	// The client request body may have been gzipped.
	body := r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		var err error
		body, err = gzip.NewReader(r.Body)
		if err != nil {
			fail500(w, "handlePostRPC", err)
			return
		}
	}

	// Prepare our Git subprocess
	cmd, pipe := gitCommand(env, "git", subCommand(rpc), "--stateless-rpc", path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fail500(w, "handlePostRPC", err)
		return
	}
	defer stdin.Close()
	if err := cmd.Start(); err != nil {
		fail500(w, "handlePostRPC", err)
		return
	}
	defer cleanUpProcessGroup(cmd) // Ensure brute force subprocess clean-up

	// Write the client request body to Git's standard input
	if _, err := io.Copy(stdin, body); err != nil {
		fail500(w, "handlePostRPC write to subprocess", err)
		return
	}

	// Start writing the response
	w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-result", rpc))
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(200) // Don't bother with HTTP 500 from this point on, just return
	if _, err := io.Copy(newWriteFlusher(w), pipe); err != nil {
		logError("handlePostRPC read from subprocess", err)
		return
	}
	if err := cmd.Wait(); err != nil {
		logError("handlePostRPC wait for subprocess", err)
		return
	}
}

func fail500(w http.ResponseWriter, context string, err error) {
	http.Error(w, "Internal server error", 500)
	logError(context, err)
}

func logError(context string, err error) {
	log.Printf("%s: %v", context, err)
}

// Git subprocess helpers
func subCommand(rpc string) string {
	return strings.TrimPrefix(rpc, "git-")
}

func gitCommand(env gitEnv, name string, args ...string) (*exec.Cmd, io.Reader) {
	cmd := exec.Command(name, args...)
	// Start the command in its own process group (nice for signalling)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Explicitly set the environment for the Git command
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("RECEIVE_APP=%s", env.App),
	)

	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	return cmd, r
}

func cleanUpProcessGroup(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}

	process := cmd.Process
	if process != nil && process.Pid > 0 {
		// Send SIGTERM to the process group of cmd
		syscall.Kill(-process.Pid, syscall.SIGTERM)
	}

	// reap our child process
	go cmd.Wait()
}

// Git HTTP line protocol functions
func pktLine(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "%04x%s", len(s)+4, s)
	return err
}

func pktFlush(w io.Writer) error {
	_, err := fmt.Fprint(w, "0000")
	return err
}

func newWriteFlusher(w http.ResponseWriter) io.Writer {
	return writeFlusher{w.(interface {
		io.Writer
		http.Flusher
	})}
}

type writeFlusher struct {
	wf interface {
		io.Writer
		http.Flusher
	}
}

func (w writeFlusher) Write(p []byte) (int, error) {
	defer w.wf.Flush()
	return w.wf.Write(p)
}

// PrereceiveHookTmpl is the template for the pre-receive hook that is created in the git repo on every push.
const PrereceiveHookTmpl = `#!/bin/bash
set -eo pipefail;
git-archive-all() {
	GIT_DIR="$(pwd)"
	cd ..
	git checkout --force --quiet $1
	git submodule --quiet update --init --recursive
	tar --create $([[ $(uname) != "Darwin" ]] && --exclude-vcs) .
}
while read oldrev newrev refname; do
	[[ $refname = "refs/heads/master" ]] && git-archive-all $newrev | {{RECEIVER}} "$RECEIVE_APP" "$newrev" "{{CACHEPATH}}" | sed -$([[ $(uname) == "Darwin" ]] && echo l || echo u) "s/^/"$'\e[1G\e[K'"/"
done
`

var prereceiveHook []byte
var cacheMtx sync.Mutex

func prepareRepo(repoRoot, cacheKey, cachePath string) (string, error) {
	cacheMtx.Lock()
	defer cacheMtx.Unlock()

	repoPath := filepath.Join(repoRoot, cacheKey)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		os.MkdirAll(repoPath, 0755)
		err = initRepo(repoPath)
		if err != nil {
			return repoPath, err
		}
	}

	cachePath, err := filepath.Abs(cachePath)
	if err != nil {
		return repoPath, err
	}

	if err := writeRepoHook(repoPath, cachePath); err != nil {
		return repoPath, err
	}

	return repoPath, nil
}

func initRepo(path string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func writeRepoHook(path, cache string) error {
	receiver := paths.SelfPath() + " receive"
	hook := strings.Replace(PrereceiveHookTmpl, "{{RECEIVER}}", receiver, 1)
	hook = strings.Replace(hook, "{{CACHEPATH}}", cache, 1)
	prereceiveHook = []byte(hook)
	return ioutil.WriteFile(filepath.Join(path, ".git", "hooks", "pre-receive"), prereceiveHook, 0755)
}
