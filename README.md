# voit
Voit file renamer.

Rename files using [Karl Voit's System][h] based on datetime in filenames.

Originally built as a pre-processor for Digikam imports to avoid regex hell in
digikam's renamer tool; this can be applied to any file.

See `voit -h` or [voit.go][i] for supported datetime formats.

## Build

``` bash
go build -o voit
./voit -h
```

## Run

``` bash
# Match default regex and rename matches.
voit -d /my/photos
> 2019-06-23T23.42.01.742 - PXL-20190623-234201742.jpg
> 2024-02-04T12.23.32.362 - PXL_20240204_122332362.jpg
voit -f /my/photos/PXL_20240204_122332362.jpg
> 2024-02-04T12.23.32.362 - PXL_20240204_122332362.jpg
voit -s -f /my/photos/PXL_20240204_122332362.jpg
> 2024-02-04T12.23.32.362.jpg

# Match using YYYYMMDDHHMMSS, lower case, and auto accept changes.
voit -l -y -p fs -d /my/photos
> 2019-06-23T23.42.01.742 - 20190623234201742.JPG
> 2024-02-04T12.23.32.362 - 20240204122332362.JPG
voit -l -y -s -p fs -d /my/photos
> 2019-06-23T23.42.01.742.jpg
> 2024-02-04T12.23.32.362.jpg

# Rename a single file, lower case, and auto accept changes.
voit -l -y -p ns -f /my/photos/signal-2024-02-04-12-23-32.JPG
> 2024-02-04T12.23.32.000 - signal-2024-02-04-12-23-32.jpg
voit -l -y -s -p ns -f /my/photos/signal-2024-02-04-12-23-32.JPG
> 2024-02-04T12.23.32.000.jpg
```

## Config
A default configuration may be specified which will be loaded when executed.
Values in this config are overridden by flags, with the exception of
**Pattern**.

``` toml
# ~/.config/voit.toml

Lower = true
Directory = "/tmp/override"

# Setting pattern ALWAYS overrides Pattern flag.
Pattern = "ns"
```

## Issues
Create a bug and provide as much information as possible.

Associate pull requests with a submitted bug.

## License
[AGPL-3.0 License][c] | [direct link][f]

## Author Information
PGP: [466EEC2B67516C7117C85CE3A0BC35D16698BAB9][d] | [github gist][e]

[c]: https://www.tldrlegal.com/license/gnu-affero-general-public-license-v3-agpl-3-0
[d]: https://keys.openpgp.org/vks/v1/by-fingerprint/466EEC2B67516C7117C85CE3A0BC35D16698BAB9
[e]: https://gist.github.com/r-pufky/a8df36977c55b5bb20829267c4c49d22
[f]: https://github.com/r-pufky/ansible_paperless_ngx/blob/main/LICENSE

[h]: https://karl-voit.at/folder-hierarchy
[i]: https://github.com/r-pufky/voit/blob/main/voit.go