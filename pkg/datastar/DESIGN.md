# DESIGN

At the moment we have datastar code in many places and examples, and its clear we need to centralise some of it. 

We do not know for sure the pattersm but telling the AI about this package means it wil start to refactor things into this package as needed

Free version we use
https://data-star.dev/reference/attributes

Pro version we might use if needed that has extra stuff.
https://data-star.dev/reference/datastar_pro

For Tailwind, i only use bun, which is installed as neeed from the pkg/dep.

The golang module we use
https://github.com/starfederation/datastar-go

The golang GUI we use built on top of datastar-go and templ.
https://github.com/CoreyCole/datastarui

---

bleve

This is really useful with datstar ? the index can be real time and later synced. Its string typed, and is really like a file based index sysrtem. Its heavily tapping into Datastar with a Broadcast system, so user get real time. We can Index deck, markdown and all the stuff we cre about. No DB or heavy stuff required. 
https://github.com/romshark/todostar/blob/main/domain/domain.go
