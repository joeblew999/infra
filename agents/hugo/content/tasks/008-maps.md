---
title: "Task 008 — Deck Maps Exploration"
summary: "Research map rendering libs and realtime geo backends for Deck."
draft: false
---

# maps

deck has a maps feature

https://github.com/ajstarks/geojson

https://github.com/ajstarks/gjc

https://github.com/ajstarks/kml

https://github.com/tidwall/tile38 can be out real time location system.

https://github.com/tidwall/tile38/pull/775 has mvt ( Mapbox Vector Tile)

https://github.com/paulmach/orb can help.

I am wondring if

## GUI system

https://github.com/ajstarks/giocanvas compiles to wasm, so we MIGHT be able to use this https://github.com/ajstarks/giocanvas/blob/master/gcdeck/main.go is the main one but its DOES NOT include all the Deck sub tools that we pipelined such that a desksh DSL file can be usd with it.

https://github.com/ajstarks/ebcanvas/blob/main/ebdeck/main.go is similar but using ebiten, which is some ways is better.
