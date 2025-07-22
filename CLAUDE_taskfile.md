# CLAUDE_taskfile.md

## File Naming Conventions

Each Taskfile must have a corresponding Markdown file using the pattern:
- `*_taskfile.yml` for the Taskfile
- `*_taskfile.md` for documentation

This ensures files are grouped together when sorted alphabetically.

## Task Naming Conventions

Use colon notation for grouped tasks: `namespace:task`, not dash notation `namespace-task`.

Examples:
- ✅ `server:create`, `server:delete` 
- ❌ `server-create`, `server-delete`

## Taskfile Tasks and Vars

### Variable Naming Convention

Use only uppercase letters for variable names

Always prefix variables with the taskfile name using double underscores (`__`) to prevent namespace clashes when including multiple taskfiles.

Pattern: `TASKFILENAME__VARIABLE_NAME`

Examples:
- `git_taskfile.yml` → `GIT__BINARY_NAME`, `GIT__DEFAULT_BRANCH`
- `hetzner_taskfile.yml` → `HETZNER__BINARY_NAME`, `HETZNER__VERSION`
- `docker_taskfile.yml` → `DOCKER__REGISTRY`, `DOCKER__IMAGE_TAG`

### Required Tasks

Always include a `vars` task that displays all taskfile variables.

Place this task immediately after the `default` task.

### Example

Example of `dummy_taskfile.yml`:
```yaml
# dummy_taskfile.yml

vars:
  DUMMY__FOO: bar
  DUMMY__VERSION: v1.0.0

tasks:
  default:
    desc: List available tasks
    cmds:
      - task --list-all

  vars:
    desc: Show taskfile variables
    cmds:
      - cmd: |
          echo "DUMMY__FOO         {{.DUMMY__FOO}}"
          echo "DUMMY__VERSION     {{.DUMMY__VERSION}}"
    silent: true
```

This prevents variable name conflicts and makes it clear which taskfile each variable belongs to.


## Taskfiles That Reference Native Binaries that are NOT downloaded

For binaries assumed to exist on all operating systems (like `git`, `docker`):

- Always include `BINARY_NAME` variable at the top of the taskfile
- Always include `BINARY_NAME_NATIVE` variable with OS-specific logic
- Always use `{{.BINARY_NAME_NATIVE}}` when calling the binary in commands

Example for git:
```yaml
vars:
  BINARY_NAME: git
  BINARY_NAME_NATIVE: '{{if eq OS "windows"}}git.exe{{else}}git{{end}}'

tasks:
  status:
    desc: Show git status
    cmds:
      - '{{.BINARY_NAME_NATIVE}} status'
```

This handles platform differences like Windows requiring `.exe` extensions.

## Taskfiles That Download Binaries

For binaries that must be downloaded from external sources:

- Always have a var like INSTALL_DIR: "{{.TASK_DIR}}/.dep" so that the highest Taskfile in the directory hierachy is used.
- Always use task name `dep` for installation
- Always include corresponding `dep:del` task for uninstallation
- Follow patterns from `hetzner_taskfile.yml` for cross-platform compatibility
- Always include these variables at the top:
  - `BINARY_NAME` - name of the binary
  - `BINARY_NAME_NATIVE` - OS/ARCH specific binary name
  - Version variable (e.g., `HCLOUD_VERSION`) - version used for download
- Always place downloaded binaries in `.dep` folder
- Always download from GitHub release tags matching the version
- Always use `{{.BINARY_NAME_NATIVE}}` when calling the binary

Example:
```yaml
vars:
  BINARY_NAME: hcloud
  HCLOUD_VERSION: v1.51.0
  BINARY_NAME_NATIVE: '{{.BINARY_NAME}}-{{OS}}-{{ARCH}}{{if eq OS "windows"}}.exe{{end}}'
```

## Syntax Rules

### Use echo, not printf

printf is not cross platform in Task files


**Wrong:**
```yaml
- printf "✅ Development server '%s' created\n" "$SERVER_NAME"
```

**Right:**
```yaml
- echo "✅ Development server '$SERVER_NAME' created"
```



### Use of Vars at start or end of a line

**Wrong:**
```yaml
- {{.BINARY_NAME_NATIVE}} context list
```

**Right:**
```yaml
- "{{.BINARY_NAME_NATIVE}} context list"
```


### Colons inside Echo statements
Never use ":" inside an echo statement.

**Wrong:**
```yaml
- echo "FOO: {{.FOO}}"
```

**Right:**
```yaml
- echo "FOO         {{.FOO}}"
```

### Colons in Descriptions
Never use `:` inside task descriptions as it breaks Taskfile parsing.

**Wrong:**
```yaml
desc: Run custom command (usage: task example:run)
```

**Right:**
```yaml
desc: Run custom command (usage - task example:run)
```

### Description Format
Task descriptions are used by parsers to generate documentation, so follow proper formatting without quotes around the description text.