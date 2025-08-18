# external

Other pkgs and project can use the same sysrtem because its all driven by just the dep.json.

We embed the dep.json, but we also look for a local one too, so if you have a dep.json we us it.


## conduit

pkg/conduit for example does this, because conduit has about 20 to 30 binaries for each platform. Its dep.json will need to be changed to use the new dep.json structure. 


## deck

pkg/deck, needs to be changed to use pkg/dep properly. 

Also i think deck also needs a version that imports all the deck code that is across many repos into a single pkg too, so we can make it run on server and as wasm. 



