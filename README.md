# voit
Voit File Renamer.

Rename files using [Karl Voit's System][h] based on datetime in filenames.

Originally built as a pre-processor for Digikam imports to avoid regex hell in
digikam; this can be applied to any file.

> [!WARNING]
> Not responsible for data loss. Always have backups and verify what is proposed
> before executing command. Though thoroughly tested and used, there still may be
> bugs that cause data loss.
>
> Always build from a released version.

## Build

``` bash
go test -v ./...
make
./bin/voit -h
```

## Run
Install in user path. Runs against current directory unless explicitly set.
Always check proposed changes before processing.

See command help for full list of options.

> [!NOTE]
> Use quotes when globbing source files. Certain shells will expand globs
> before passing arguments to binaries resulting in unexpected behavior.
> Wrapping the path in quotes prevents the user shell from expanding these.

### Rename
``` bash
# Most photos today include milliseconds, so use --photo-ms
$ voit rename --photo
> 2017-02-25.jpg                   ➔ 2017-02-25T00.00.00.000 2017-02-25.jpg
> download_20170419_134641.jpg     ➔ 2017-04-19T13.46.41.000 download_20170419_134641.jpg
> download_20170419_134641.jpg.xmp ➔ 2017-04-19T13.46.41.000 download_20170419_134641.jpg.xmp
> ...
> Proposed changes: 2602 file(s).
> Proceed? (y/n): y
> Renamed in 911.35377ms.

# Target a single file.
voit rename --photo-ms -s /my/photos/PXL_20240204_122332362.jpg

voit rename --photo-ms -s /my/other/photos  # Target directory.
voit rename --photo-ms -s "/globbing/supported/PXL_2024*.jpg"  # Glob files using qoutes.

# Remove description.
voit rename --no-desc -s /my/photos/PXL_20240204_122332362.jpg
> 2024-02-04T12.23.32.362.jpg

# Rename a subset of unix datestamps.
$ voit rename --unix --no-desc -s "./144*"
1441569996446.jpg     ➔ 2015-09-06T20.06.36.446.jpg
1441569996446.jpg.xmp ➔ 2015-09-06T20.06.36.446.jpg.xmp
```

### Tag
``` bash
# Set all files explicitly to 'party' tags.
voit tag -e party

# Add 'candles', 'bday' tags to all files with 'cake' tag.
voit tag -c cake -a candles -a bday

# Remove 'candles', 'bday' tags.
voit tag -r candles -r bday  # From all files.
voit tag -r candles -r bday -l cake  # From all files with 'cake'.

# Remove all tags.
voit tag -d  # From all files.
voit tag -d -c candles bday  # From all files with 'candles', 'bday'.

# Globbing supported. Use quotes.
voit tag -s "./photos/PXL_2024*.jpg" -a candles -a bday

# Sync XMP tags to both file and sidecar.
voit tag --sync-xmp
```

## Config
A default configuration may be specified which will be loaded when executed.
Values in this config are superseded by flags.

### ~/.config/voit.toml
``` toml
# Only define options that are preferred defaults.

# Global Options
Yes = true
Verbose = true
TagSep = " = "
DescSep = " - "
SpanSep = "--"

[Rename]
no-desc = false

[Tag]
add = ['summer', 'beach', 'vacation']
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
