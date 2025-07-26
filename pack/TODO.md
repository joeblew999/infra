# TODO

AI: this file is a game plan for what we need to do:


Gameplan:

ONLY do each step on it own.

   1. pkg/pack/npm needs to have the code we have in main.go. 

   2. pkg/config needs most of the constants we have in main.go. make sure pkg/pack/npm uses them

   3. pkg/dep needs to install the gh cli. make sure pkg/pack/npm uses them.

   4. Move the `.env` variables to root.


AFTER

The existing cli command "build", can then incorporate all this, because it a workflow.

- I think we need to revisit the naming of all the cli commamnds. We are exposing too many commands. We want idempotent and that what workflows do, like build, release, deploy.   cli command "build" does to much ?

Maybe we ONLY need build, test, release ?? 

