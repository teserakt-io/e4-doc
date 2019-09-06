
# E4 Documentation

This repository holds documentation for E4. It's based on [hugo](https://gohugo.io/).


## Development

Clone the repository then run
```
git submodule update # pull the theme
hugo serve -D # start local hugo server on localhost:1313, including draft documents
```

Hugo will serve the website locally, and live reload when changes are made.
Only when files get deleted, the build folder need to be deleted
```
rm -rf public/
```
