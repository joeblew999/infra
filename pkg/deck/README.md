# deck

Deck is a system that can output to SVG and so stream into DataStar

Users can edit the XML document, and its system will produce an SVG.

The SVG will be later used with DataStar, so we can easily stream the output to a Browser. 


## Decksh

https://github.com/ajstarks/decksh/tree/master/cmd has 3 command to help with the decksh system 

Takes a Decksh and coverts it to DeckXml-


https://github.com/ajstarks/decksh/blob/master/test.dsh is an example Decksh file.
 
https://github.com/ajstarks/decksh/blob/master/test.xml is the output as DeckXml file.


## Deck Exports

https://github.com/ajstarks/deck

Takes a DeckXml and concert it to various oututs.

PDF: https://github.com/ajstarks/deck/tree/master/cmd/pdfdeck
PNG: https://github.com/ajstarks/deck/tree/master/cmd/pngdeck
SVG: https://github.com/ajstarks/deck/tree/master/cmd/svgdeck


## Web

Very basic controllers..

https://github.com/ajstarks/deck/tree/master/cmd/deckd

https://github.com/ajstarks/deck/tree/master/cmd/deckweb


## GEO

https://github.com/ajstarks/kml/blob/master/cmd/geodeck/main.go is for making 2d maps by Convert KML files to deck markup.

## Image Conversions

https://github.com/ajstarks/giftsh

Takes a giftsh, and does transformations on images.

Can be used alognj with the other things above or after.

## Examples

https://github.com/ajstarks/dubois-data-portraits

https://github.com/ajstarks/deckviz

NOT that it seems to use DECKFONTS as an env or flag to help it location the fonts that are needed. This is a repo of them here at https://github.com/ajstarks/deckfonts, but i think we will need a better way of getting Fonts off Google fonts ? 



