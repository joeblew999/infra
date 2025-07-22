# Atlantis

Code: https://github.com/runatlantis/atlantis

Version; https://github.com/runatlantis/atlantis/releases/tag/v0.35.0

Docs: https://www.runatlantis.io/docs.html

Terraform Pull Request Automation, ruuns Terraform Workflows with web hooks.

It will run terraform for you and has web hooks.

So I can just feed them to NATS.

All binaries running can also do the same , so so do not need nats.go in my binaries.

NATS can record all repositories, where the docker images live ( github ), and where the dockers are running , and all stats so it can auto scale any docker on any cloud using nats with terraform.

Datastar web gui to see it all :)

Credentials to each cloud are stored in nats as secrets and supplied. https://www.runatlantis.io/docs/provider-credentials.html

Goss might help: https://github.com/goss-org/goss as they also use it.


Need to make a simple golang package that sends web hooks, so that all binaries tell NATS there stats using github.com/shirou/gopsutil/v4 from https://github.com/shirou/gopsutil

https://claude.ai/chat/311bc30f-8e9d-4b52-8028-5b96f2aa0db8 is the plan for it.





