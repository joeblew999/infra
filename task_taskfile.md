# task 

https://taskfile.dev

https://taskfile.dev/next/reference/templating/

builtins

Taskfile built-in variables provide useful information about the current task's execution environment. Here are the common built-ins you can use:

### Task Information

* **`{{.TASK}}`**: The name of the current task. For example, if you run `task build`, this variable will be `build`.
* **`{{.TASK_DIR}}`**: The directory where the `Taskfile.yml` defining the current task is located. This is useful when your Taskfile might be in a different directory than where you execute `task`.
* **`{{.TASK_FILE}}`**: The absolute path to the `Taskfile.yml` defining the current task.

***

### Environment and System Information

* **`{{.OS}}`**: The operating system name (e.g., `linux`, `darwin` for macOS, `windows`).
* **`{{.ARCH}}`**: The system architecture (e.g., `amd64`, `arm64`).
* **`{{.USER_WORKING_DIR}}`**: The directory from which `task` was invoked by the user. This is particularly useful if your Taskfile is in a different location but you need to reference files relative to the user's current directory.
* **`{{.CLI_ARGS}}`**: All arguments passed to `task` after the task name, as a single string. For example, if you run `task run --debug some_file.txt`, this would be `--debug some_file.txt`.
* **`{{.CLI_ARGS_ARRAY}}`**: All arguments passed to `task` after the task name, as an array of strings. Using the previous example, this would be `["--debug", "some_file.txt"]`.
* **`{{.GO_ARCH}}`**: The Go architecture (e.g., `amd64`, `arm64`). This is generally the same as `.ARCH` but specifically refers to the Go environment.
* **`{{.GO_OS}}`**: The Go operating system (e.g., `linux`, `darwin`, `windows`). Similar to `.OS`.

***

### Utility Functions

Task also provides built-in functions that can be used within variable expressions for common operations:

* **`{{.SH_QUOTE <string>}}`**: Quotes a string for safe use in shell commands, handling spaces and special characters. This is crucial when passing user-defined variables or file paths that might contain spaces.
* **`{{.YAML_PARSE <yaml_string>}}`**: Parses a YAML string and makes its contents accessible.
* **`{{.JSON_PARSE <json_string>}}`**: Parses a JSON string and makes its contents accessible.
* **`{{.BASE <path>}}`**: Returns the last element of a path (the filename or directory name).
* **`{{.DIR <path>}}`**: Returns all but the last element of a path (the directory name).
* **`{{.EXT <path>}}`**: Returns the file extension of a path.
* **`{{.NOEXT <path>}}`**: Returns the path without its file extension.
* **`{{.NOTDIR <path>}}`**: Returns the filename without its directory path.
* **`{{.ABS <path>}}`**: Returns the absolute path of a given path.
* **`{{.REL <base_path> <target_path>}}`**: Returns the relative path from `base_path` to `target_path`.
* **`{{.FILE_EXISTS <path>}}`**: Returns `true` if the file exists, `false` otherwise.
* **`{{.DIR_EXISTS <path>}}`**: Returns `true` if the directory exists, `false` otherwise.
* **`{{.ARCHIVE_ENTRY_EXISTS <archive_path> <entry_path>}}`**: Checks if an entry exists within a zip or tar archive.
* **`{{.SEQ <start> <end>}}`**: Generates a sequence of numbers from `start` to `end`.
* **`{{.CAT <file_path>}}`**: Reads the content of a file.
* **`{{.FROM_JSON <string>}}`**: Decodes a JSON string.
* **`{{.TO_JSON <value>}}`**: Encodes a value to a JSON string.
* **`{{.FROM_YAML <string>}}`**: Decodes a YAML string.
* **`{{.TO_YAML <value>}}`**: Encodes a value to a YAML string.
* **`{{.KEYS <map>}}`**: Returns the keys of a map.
* **`{{.VALUES <map>}}`**: Returns the values of a map.
* **`{{.INDENT <spaces> <string>}}`**: Indents a multi-line string.
* **`{{.TRIM <string>}}`**: Removes leading and trailing whitespace from a string.
* **`{{.SPRINTF <format> <args>...}}`**: Formats a string using `fmt.Sprintf` syntax.
* **`{{.TOLOWER <string>}}`**: Converts a string to lowercase.
* **`{{.TOUPPER <string>}}`**: Converts a string to uppercase.
* **`{{.REPLACE <string> <old> <new> <n>}}`**: Replaces occurrences of `old` with `new` in a string, up to `n` times.
* **`{{.SPLIT <string> <sep>}}`**: Splits a string by a separator.
* **`{{.HAS_PREFIX <string> <prefix>}}`**: Checks if a string has a given prefix.
* **`{{.HAS_SUFFIX <string> <suffix>}}`**: Checks if a string has a given suffix.
* **`{{.HTTP_GET <url>}}`**: Performs an HTTP GET request and returns the response body.
* **`{{.HTTP_STATUS <url>}}`**: Performs an HTTP GET request and returns the status code.
* **`{{.HASH <string>}}`**: Calculates the SHA256 hash of a string.
* **`{{.UUID}}`**: Generates a new UUID.
* **`{{.NOW}}`**: Returns the current time.
* **`{{.ENV <variable>}}`**: Returns the value of an environment variable.
* **`{{.GIT_BRANCH}}`**: Returns the current Git branch name.
* **`{{.GIT_LAST_TAG}}`**: Returns the last Git tag.
* **`{{.GIT_LAST_TAG_OR_COMMIT}}`**: Returns the last Git tag or commit hash.
* **`{{.GIT_IS_DIRTY}}`**: Returns `true` if the Git repository is dirty.
* **`{{.GIT_IS_TAGGED}}`**: Returns `true` if the current commit is tagged.
* **`{{.GIT_COMMIT}}`**: Returns the current Git commit hash.
* **`{{.GIT_COMMIT_SHORT}}`**: Returns the short Git commit hash.

These built-ins allow for dynamic and flexible Taskfile configurations.
