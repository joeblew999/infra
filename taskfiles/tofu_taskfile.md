# Generic OpenTofu Taskfile

This Taskfile provides a generic set of commands for managing OpenTofu projects. It is designed to be provider-agnostic.

## Basic Usage

Here's a simple workflow to initialize a project, review the plan, and apply the changes.

1.  **Initialize your OpenTofu project:**
    ```bash
    task init
    ```

2.  **Create an execution plan:**
    ```bash
    task plan
    ```

3.  **Apply the changes:**
    ```bash
    task apply
    ```

You can override the default terraform directory (`./terraform`) by setting the `TF_DIR` variable:

```bash
task plan TF_DIR=./my-other-project/
```