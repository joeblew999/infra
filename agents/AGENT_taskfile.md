# CLAUDE_taskfile.md

All the reusable taskfiles live in the taskfiles folder.

Reusable taskfiles start with the NAME_taskfile.yml pattern.
Non reusable ones, as per Taskfile conventions, are called Taskfile.yml

The taskfiles are designed to be included using remote taskfiles feature.

## Useful flags when calling taskfiles

```sh
# List tasks as JSON
task --list --json

# CLI_ARGS_LIST array variable which contains the arguments passed to Task after the -- (the same as CLI_ARGS, but an array instead of a string).
task --CLI_ARGS_LIST=for,bar vars 

# CLI_ARGS usage
task CLI_ARGS=for vars 


```


## AI agents

The following Taskfiles are designed to help configure AI Agents:

- gemini_taskfile.yml
- claude_taskfile.yml

Unless an exclusion is in a Taskfile, this is the way to code taskfiles.

## Best pratice example

dummy_taskfile.yml is our best practice reference example.

Testing directly:
```sh
task -d taskfiles/dummy_taskfile.yml --list-all

task -d taskfiles/dummy_taskfile.yml vars
```

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

All taskfiles should have a "vars" task for pinting the vars. This is to help with discovery and debuggings.

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

- Always include `DUMMY__BINARY_NAME` variable at the top of the taskfile
- Always include `DUMMY__BINARY_NAME_NATIVE` variable with OS-specific naming
- Always use `{{.DUMMY__BINARY_NAME_NATIVE}}` when calling the binary in commands

Example for git:
```yaml
vars:
  DUMMY__BINARY_NAME: git
  DUMMY__BINARY_NAME_NATIVE: '{{if eq OS "windows"}}git.exe{{else}}git{{end}}'

tasks:
  status:
    desc: Show git status
    cmds:
      - '{{.DUMMY__BINARY_NAME_NATIVE}} status'
```

This handles platform differences like Windows requiring `.exe` extensions.

## Taskfiles That Download Binaries

For binaries that must be downloaded from external sources:

- Always include these variables at the top:
  - `DUMMY__BINARY_NAME` - name of the binary
  - `DUMMY__BINARY_NAME_NATIVE` - OS/ARCH specific binary name that is used in the taskfile to call the binary
  - `DUMMY__VERSION` - version of the binary downloaded from GitHub releases
  - `DUMMY__INSTALL_DIR: "{{.TASK_DIR}}/.dep"` - location of the final binary on disk
- Always use `{{.DUMMY__INSTALL_DIR}}/tmp` for temporary downloading and unpacking operations
- Always use the task name `dep` for the binary download, following `dummy_taskfile.yml` for cross-platform compatibility with a separate sub command per OS where needed
- Always include corresponding `dep:del` task for uninstallation


Example:
```yaml
vars:
  HCLOUD__NAME: hcloud
  HCLOUD__VERSION: v1.51.0
  HCLOUD__BINARY_NAME_NATIVE: '{{.HCLOUD__NAME}}-{{OS}}-{{ARCH}}{{if eq OS "windows"}}.exe{{end}}'
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