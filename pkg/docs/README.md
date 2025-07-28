# docs

This pkg is currently design markdown IN and spits out HTML.

Currently is dynamic.

Plan is to allow it to work off any folder that is the docs root.

we then any github repo with a .docs in it is well known to the system.
We might have to introduce the concept of a spurce reposity though, so that anyone using this, can configure a repo and Infra will pull it and use the docs inside the repo. Lets first wprk it ut properly though.

i suspect we need to get caddy into the system first, so that on FLY caddy is running . CADDY is part of the pkg/cmd but i am not sure its right for it to be there.



