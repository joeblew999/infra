# web

Using https://github.com/delaneyj/toolbelt/tree/main/natsrpc for the protos. It has an Example.
- This Allows us to define Types usign Protobufs, so that we can support Schema Evolution at runtime.

Using https://github.com/delaneyj/toolbelt/tree/main/embeddednats for the Embedded NATS Server using golang.
- This allows us to run an embedded NATS Server as a Leaf Node, allowing a Global Nats Server to work with the Embeded NATS Leaf node.

Using https://github.com/starfederation/datastar-go for the DataStar using golang.
- I not not have a datastar-go CLAUDE.md yet
- This is the DataStar system.  https://data-star.dev
- Docs: https://data-star.dev/reference/attributes

Using: https://github.com/CoreyCole/datastarui for the DataStar GUI using golang.
- Claude is using ./../CLAUDE_datastarui.md
- Has an MPC for playwright. 
- Docs: https://datastar-ui.com/docs

## VS Code Extensions

Templ: https://marketplace.visualstudio.com/items?itemName=a-h.templ
Datastar: https://marketplace.visualstudio.com/items?itemName=starfederation.datastar-vscode



## Architetcural flow.

Feeds data into NATS --> DataStar --> DataStarUI --> Browser.


