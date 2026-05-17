package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// resolves the target path and returns the absolute path
// and a sorted list of all entries, where directories are marked with "/"
// including files inside subdirectories
func resolveAndReadDir() (string, []string) {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	var entries []string
	filepath.WalkDir(abs, func(path string, dir os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(abs, path)
		if rel == "." {
			return nil
		}
		if strings.HasPrefix(dir.Name(), ".") {
			if dir.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if dir.IsDir() {
			rel += "/"
		}
		entries = append(entries, rel)
		return nil
	})

	sort.Strings(entries)
	return abs, entries
}

// opens the file in Helix and waits for the editor to close
func launchEditor(path string) {
	editor := "hx"
	_, err := exec.LookPath("hx")
	if err != nil {
		_, err := exec.LookPath("helix")
		if err == nil {
			editor = "helix"
		}
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

// creates a temp file and writes lines into it
func writeTempFile(lines []string) *os.File {
	tmp, err := os.CreateTemp("", "bufdir_*.txt")
	if err != nil {
		panic(err)
	}
	for _, line := range lines {
		fmt.Fprintln(tmp, line)
	}
	tmp.Close()
	return tmp
}

// reads non-empty trimmed lines from a file
func readLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// diffs before and after to produce creations, deletions and renames
func computeChanges(before, after []string) (creates, deletes []string, renames [][2]string) {
	beforeMap := make(map[string]bool)
	afterMap := make(map[string]bool)

	for _, v := range before {
		beforeMap[v] = true
	}

	for _, v := range after {
		afterMap[v] = true
	}

	for _, v := range before {
		if !afterMap[v] {
			deletes = append(deletes, v)
		}
	}

	for _, v := range after {
		if !beforeMap[v] {
			creates = append(creates, v)
		}
	}

	// pair deletes and creates positionally as renames
	// (up to min(len(creates), len(deletes)))
	n := len(creates)
	if len(deletes) < n {
		n = len(deletes)
	}

	for i := 0; i < n; i++ {
		renames = append(renames, [2]string{deletes[i], creates[i]})
	}

	creates = creates[n:]
	deletes = deletes[n:]
	return
}

// creates the listed files and directories
func applyCreates(abs string, names []string) {
	for _, name := range names {
		p := filepath.Join(abs, stripDir(name))

		_, err := os.Stat(p)
		if err == nil {
			continue
		}

		if strings.HasSuffix(name, "/") {
			err = os.MkdirAll(p, 0755)
			if err != nil {
				fmt.Printf("mkdir  %s FAILED: %v\n", name, err)
			} else {
				fmt.Printf("mkdir  %s\n", name)
			}
		} else {
			ensureParent(p)

			f, err := os.Create(p)
			if err != nil {
				fmt.Printf("touch  %s FAILED: %v\n", name, err)
			} else {
				f.Close()
				fmt.Printf("touch  %s\n", name)
			}
		}
	}
}

// removes the listed files and directories
func applyDeletes(abs string, names []string) {
	for _, name := range names {
		p := filepath.Join(abs, stripDir(name))

		info, err := os.Stat(p)
		if err != nil {
			continue
		}

		if info.IsDir() {
			err := os.RemoveAll(p)
			if err != nil {
				fmt.Printf("rmdir  %s FAILED: %v\n", name, err)
			} else {
				fmt.Printf("rmdir  %s\n", name)
			}
		} else {
			os.Remove(p)
			fmt.Printf("rm  %v\n", name)
		}
	}
}

// renames files and directories according to the given pairs
func applyRenames(abs string, pairs [][2]string) {
	for _, pair := range pairs {
		src := filepath.Join(abs, stripDir(pair[0]))
		dist := filepath.Join(abs, stripDir(pair[1]))

		_, err := os.Stat(src)
		if err != nil {
			continue
		}

		ensureParent(dist)

		err = os.Rename(src, dist)
		if err != nil {
			fmt.Printf("mv  %s -> %s FAILED: %v\n", pair[0], pair[1], err)
		} else {
			fmt.Printf("mv  %s -> %s\n", pair[0], pair[1])
		}
	}
}

// removes trailing "/" from a name
func stripDir(name string) string {
	return strings.TrimSuffix(name, "/")
}

// creates parent directories for the given file path if needed
func ensureParent(path string) {
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		os.MkdirAll(dir, 0755)
	}
}

func main() {
	abs, before := resolveAndReadDir()

	tmp := writeTempFile(before)
	defer os.Remove(tmp.Name())

	launchEditor(tmp.Name())

	after := readLines(tmp.Name())

	creates, deletes, renames := computeChanges(before, after)

	applyCreates(abs, creates)
	applyDeletes(abs, deletes)
	applyRenames(abs, renames)
}
