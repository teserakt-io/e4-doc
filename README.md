
# E4 Documentation

This repository holds documentation for E4. It's based on [hugo](https://gohugo.io/).

## Development

First, install Hugo:

``` bash
go get -u github.com/gohugoio/hugo
go install -a --tags extended github.com/gohugoio/hugo
```

Clone the repository then run:

```
git submodule update # pull the theme
hugo serve -D # start local hugo server on localhost:1313, including draft documents
```

Hugo will serve the website locally, and live reload when changes are made.
Only when files get deleted, the build folder need to be deleted
```
rm -rf public/
```

# Publish to github pages

To publish to https://teserakt-io.github.io/e4-doc/, commit your changes and run

```
./publish.sh
```

A special `gh-pages` branch is set to track the content of the public folder. The script will generate it to the latest version, and commit the public/ content to this branch. Once pushed, github will automatically deploy it and it will be available after a short time.
