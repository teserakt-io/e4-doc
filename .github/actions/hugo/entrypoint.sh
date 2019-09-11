#!/bin/bash

set -e

# Pull the theme
git submodule update --init --recursive

hugo $*
