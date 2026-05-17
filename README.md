# bufdir

```bash
go install github.com/ifdiego/bufdir@latest
```

Edit your directory like a text file.

Open a directory with `bufdir` (or `bufdir <directory>`), edit the
listing files in Helix, save and quit — and bufdir applies your
changes: create, delete, rename, and move files and directories.

What you do / What happens

* Add a new line `foo.txt` → file is created (empty)
* Add a line ending with `/` → directory is created
* Delete a line → file or directory is removed
* Change a filename → file or directory is renamed
* Change `sub/foo.txt` to `foo.txt` → file is moved to the root
* Change `foo.txt` to `sub/foo.txt` → file is moved into a subdirectory

Hidden files and directories (names starting with `.`) are ignored.

Similar to what [oil.nvim](https://github.com/stevearc/oil.nvim) does.
