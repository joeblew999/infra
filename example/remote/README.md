# examples

If your using this from your own repo you will want to use Remote Taskfiles.

## Remote task files usage

You can reference Taskfiels from anywhere: https://taskfile.dev/experiments/remote-taskfiles/

In the example Taskfile, we are using "joeblew999/infra", so that it calls this actual repo.

WHen you tag ypour repo, you can then reference by tags also.
https://github.com/joeblew999/infra/releases/tag/v1.0.0

Here are some usage examples:

```sh

# TO allow remote taskfiles and not prompt you.
TASK_X_REMOTE_TASKFILES=1 task --yes


# Cache management:
# Clear cache. The .task folder will be cleaned.
TASK_X_REMOTE_TASKFILES=1 task --clear-cache

# Force refresh of remote includes. The .task folder will be populated.
TASK_X_REMOTE_TASKFILES=1 task --yes --download


# List the tasks available.
TASK_X_REMOTE_TASKFILES=1 task --yes --list-all --sort alphanumeric

# There is "task" taskfile to help with debugging, so lets play with that:

# Dry run, if you are not sure you trust it yet
TASK_X_REMOTE_TASKFILES=1 task --yes --dry task:vars
# Dry run, and want to see everything it does.
TASK_X_REMOTE_TASKFILES=1 task --yes --dry --verbose task:vars



# Get its vars to help with debugging.
TASK_X_REMOTE_TASKFILES=1 task --yes task:vars

# Call dummy too.
TASK_X_REMOTE_TASKFILES=1 task --yes dummy:vars

# Call git 
TASK_X_REMOTE_TASKFILES=1 task --yes git:vars

# List all git tasks
TASK_X_REMOTE_TASKFILES=1 task --yes git --list
# List all git tasks as json
TASK_X_REMOTE_TASKFILES=1 task --yes git --list --json

# Bypass your local root Taskfile, to list operations on a specific taskfile.
TASK_X_REMOTE_TASKFILES=1 task --yes --taskfile https://raw.githubusercontent.com/joeblew999/infra/main/taskfiles/git_taskfile.yml --list

# Bypass your local root Taskfile, to list operations on a specific root taskfile.
TASK_X_REMOTE_TASKFILES=1 task --yes --taskfile https://raw.githubusercontent.com/joeblew999/infra/main/example/remote/Taskfile.yml --list


```