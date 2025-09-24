# 005-fs-sync

we want our users to be able to bring in a git repo at runtime.

this is their "project".

But we want our Server to be serverless. Just like how we use pocketbase and jetstream, so we can boot a server, and it will get the Pocketbase DB restored off S3, while also keeping it synced to S3.

So we need a way to do that with git repos. I can imagine that we might store what git repos a usr has in Pocketbase too.



https://github.com/go-git/go-git


https://github.com/restic/restic

https://github.com/restic/restic/releases/tag/v0.18.1

