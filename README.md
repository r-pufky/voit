# voit
Voit file renamer.

Rename files using [Karl Voit's System][h] based on datetime in filenames.

Originally built as a pre-processor for Digikam imports to avoid regex hell in
digikam; this can be applied to any file.

> [!WARNING]
> Not responsible for data loss. Always have backups and verify what is proposed
> before executing command. Though thoroughly tested and used, there still may be
> bugs that cause data loss.

## Build

``` bash
go test ./... -v
make  # manually build development with: go build -o voit
./voit -h
./voit rename -h
```

## Run

``` bash
# Rename typical camera photos, one with existing tags
voit rename --photo-ms -s /my/photos
> 2019-06-23T23.42.01.742 PXL-20190623-234201742.jpg
> 2024-02-04T12.23.32.362 PXL_20240204_122332362 -- fruit.jpg

# Individual files may be targetted too.
voit rename --photo-ms -s /my/photos/PXL_20240204_122332362.jpg
> 2024-02-04T12.23.32.362 PXL_20240204_122332362.jpg

# Strip existing date pattern from description.
voit rename --photo-ms --strip -s /my/photos/PXL_20240204_122332362.jpg
> 2024-02-04T12.23.32.362 PXL_.jpg

# Remove description.
voit rename --no-desc -s /my/photos/PXL_20240204_122332362.jpg
> 2024-02-04T12.23.32.362.jpg

# Use alternative separators, remove description and tags.
voit rename --desc-sep " - " --tag-sep " = " --photo-ms --no-desc --no-tags -s /my/photos/2024-02-04T12.23.32.362 - PXL_20240204_122332362 = tacos.jpg
> 2024-02-04T12.23.32.362.jpg
```

## Config
A default configuration may be specified which will be loaded when executed.
Values in this config are overridden by flags, with the exception of
**Pattern**.

``` toml
# ~/.config/voit.toml

# Only define options that should be set.

# Global Options
Yes       = true
Verbose   = true
TagSep    = " = "
DescSep   = " - "
SpanSep   = "--"

# Rename Options
[Rename]
# Setting pattern ALWAYS overrides Pattern flag.
Pattern       = 'photo-ms'
Lower         = true
Strip         = false
NoDesc        = false
NoTags        = false
Overwrite     = true
PreferPattern = true
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