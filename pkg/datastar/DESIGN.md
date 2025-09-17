# DESIGN

At the moment we have datastar code in many places and examples, and its clear we need to centralise some of it. 


Free version we use
https://data-star.dev/reference/attributes

Pro version we might use if needed that has extra stuff.
https://data-star.dev/reference/datastar_pro

For Tailwind, i only use bun, which is installed as neeed from the pkg/dep.

The golang module we use
https://github.com/starfederation/datastar-go

The golang GUI we use built on top of datastar-go and templ.
https://github.com/CoreyCole/datastarui


- https://templ.guide
- https://templ.guide/developer-tools/cli
- go install github.com/a-h/templ/cmd/templ@latest, so we probably need to use pkg/dep to make sure its installed, like we do for many things ?
- so we need templ and go generate ./... integrated to use it. We already have some code generation system called PreFlight i think ? 




---

bleve

This is really useful with datstar ? the index can be real time and later synced. Its string typed, and is really like a file based index sysrtem. Its heavily tapping into Datastar with a Broadcast system, so user get real time. We can Index deck, markdown and all the stuff we cre about. No DB or heavy stuff required. 

https://github.com/romshark/todostar/blob/main/domain/domain.go

Its uses https://webawesome.com for the html components, which is maybe better..
