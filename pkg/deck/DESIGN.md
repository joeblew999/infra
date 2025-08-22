# Design

Deck and decksh is a DSL to generate outputs in different formats.

Core code we will use:

- https://github.com/ajstarks/decksh is the DSL with 3 cmds's, that produces Deck XML

/Users/apple/workspace/go/src/github.com/joeblew999/infra/pkg/deck/.source/decksh/cmd/ has the 3 cmds, that we need as pkg.


- https://github.com/ajstarks/deck are the Outputters that outout to SVG, PNG, PDF that we need as pkg

PNG REF CODE:
/Users/apple/workspace/go/src/github.com/joeblew999/infra/pkg/deck/.source/deck/cmd/pngdeck/

PDF ref code:
/Users/apple/workspace/go/src/github.com/joeblew999/infra/pkg/deck/.source/deck/cmd/pdfdeck/

SVG Ref Code:
/Users/apple/workspace/go/src/github.com/joeblew999/infra/pkg/deck/.source/deck/cmd/svgdeck/

- its a bummer that all the logic is in the cmd folders, because we cant import it, but must copy it and use it.

Examples we can use:
- https://github.com/ajstarks/dubois-data-portraits - examples
- https://github.com/ajstarks/deckviz - examples


need fonts from pkg/font package in this repo.

The main thing we are interessted in is the SVG output, becaue we can stream that through using Datastar to have a real time updating Web GUI.

1. Make golang code that imports all the parts of Deck that we want so that we can control it without needing to pipeline it using many binaries or wasm.

2. Make a little prototype, in this pkg using a go.work, that use our deck pkg and datastar to show it working, and responding to changes in the Deck DSL.

3. Make an Editor, that uses https://github.com/CoreyCole/datastarui, to allow editing deck DSL in the Web. How we do this will for the first version:

- have 3 panes:

- top pane: Output view 

- left pane: File View of the Deck files on the Server File System, allowing Usrs to move between what decks they want to edit.

- bottom pane: as a Usr clicks on elements ion the Output view, we show an Editor, using the datastarui Sheets feature. 

