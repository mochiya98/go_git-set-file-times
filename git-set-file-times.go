package main

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func main() {
	zeroBytes := []byte{0}
	files := make(map[string]struct{})

	cmd := exec.Command("git", "ls-files", "-z")
	var err error
	stdout, _ := cmd.StdoutPipe()
	rd := bufio.NewReader(stdout)
	cmd.Start()
	for line := []byte{}; err == nil; line, err = rd.ReadSlice(0) {
		if len(line) > 1 {
			files[string(line[:len(line)-1])] = struct{}{}
		}
	}
	cmd.Wait()

	var mtime time.Time
	cmd = exec.Command("git", "--no-pager", "log", "-m", "-r", "--name-only", "--no-color", "--pretty=raw", "-z")
	out, _ := cmd.Output()
	lines := bytes.Split(out, []byte{'\n'})
	reCommitTime := regexp.MustCompile(`^committer .*? (\d+) (?:[\-\+]\d+)$`)
	reCommitFiles := regexp.MustCompile(`\0\0commit [a-f0-9]{40}(?: \(from [a-f0-9]{40}\))?$`)
	for _, line := range lines {
		if g := reCommitTime.FindSubmatch(line); g != nil {
			sec, _ := strconv.Atoi(string(g[1]))
			mtime = time.Unix(int64(sec), 0)
		} else if reCommitFiles.Match(line) || bytes.HasSuffix(line, zeroBytes) {
			fns := bytes.Split(line, zeroBytes)
			for _, fn := range fns {
				strFn := string(fn)
				if _, ok := files[strFn]; ok {
					delete(files, strFn)
					os.Chtimes(strFn, mtime, mtime)
				}
			}
		}
		if len(lines) == 0 {
			break
		}
	}
}
