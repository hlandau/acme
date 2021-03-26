# How to download binary releases programmatically

## With curl

```sh
VER="$(curl -s -H 'Accept: application/vnd.github.v3+json' 'https://api.github.com/repos/hlandau/acmetool/releases/latest' | python -c 'import sys,json;k=json.load(sys.stdin);print(k["tag_name"])')"
curl -Ls -o acmetool-bin.tar.gz "https://github.com/hlandau/acmetool/releases/download/$VER/acmetool-$VER-linux_amd64_cgo.tar.gz"
```

## With wget

```sh
VER="$(wget --quiet -O - --header='Accept: application/vnd.github.v3+json' 'https://api.github.com/repos/hlandau/acmetool/releases/latest' | python -c 'import sys,json;k=json.load(sys.stdin);print(k["tag_name"])')"
wget --quiet -O acmetool-bin.tar.gz "https://github.com/hlandau/acmetool/releases/download/$VER/acmetool-$VER-linux_amd64_cgo.tar.gz"
```
