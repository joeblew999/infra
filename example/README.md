# examples

If your using this from your own repo you will want to use Remtoe Taskfiles.

## Remote task files usage

You can reference Taskfiels from anywhere: https://taskfile.dev/experiments/remote-taskfiles/

In the example Taskfile, we are using "joeblew999/infra", so that it calls this actual repo.

Here are some usage examples:

```sh

# TO allow remote taskfiles and not prompt you.
TASK_X_REMOTE_TASKFILES=1 task --yes


# With sorting for easier reading.
TASK_X_REMOTE_TASKFILES=1 task --yes --sort alphanumeric

# call the task Taskfile to get its vars to help with debugging.
TASK_X_REMOTE_TASKFILES=1 task --yes task:vars

# call the dummy Taskfile to get its vars.
TASK_X_REMOTE_TASKFILES=1 task --yes dummy:vars

# call the git Taskfile to get its vars.
TASK_X_REMOTE_TASKFILES=1 task --yes git:vars

# list all tasks it has as json
TASK_X_REMOTE_TASKFILES=1 task --yes git --list
# list all tasks it has as json
TASK_X_REMOTE_TASKFILES=1 task --yes git --list --json

# Bypass your Taskfile, to list operations on a specific taskfile.

TASK_X_REMOTE_TASKFILES=1 task --taskfile https://raw.githubusercontent.com/joeblew999/infra/main/taskfiles/git_taskfile.yml --list

# 
TASK_X_REMOTE_TASKFILES=1 task --taskfile https://raw.githubusercontent.com/joeblew999/infra/main/example/Taskfile.yml --list


```