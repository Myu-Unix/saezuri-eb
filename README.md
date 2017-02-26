## Saezuri-eb

an unfinished Go toy implementation of a GNU Social client made with Golang and ebiten (https://hajimehoshi.github.io/ebiten) 

![Gif](images/output.gif)

#### What is it good for ?

It's an unfinished toy client so don't expect it to be useful. It works well to quickly glance at timelines and will refresh itself automatically.

#### What is it bad for ?

Everything not listed in *What is it good for ?* :)

#### Quickstart

You need at these prerequisites :

+ Golang
+ Linux with GLX driver.

clone the repository then create a file called saezuri.conf within the project folder and fill it with your login details :

	username
	password
	instance_url (e.g https://quitter.se)

Within the source directory, build & launch saezuri-eb :

    export GOPATH=/home/username/go/
    cd /home/username/go/saezuri-eb
    go get ./...
    go build saezuri-eb.go
    ./saezuri-eb