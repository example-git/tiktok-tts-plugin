![Haunt](assets/haunt-wordmark.png)

# TikTok TTS server extension

A standalone Haunt speech extension. The package runs on the server and supplies
its configuration fields and voice catalog to browser and desktop clients through
Haunt's dynamic menus. It uses the Go standard library.

## Build and install

The **Extension packages** GitHub Actions workflow tests the provider and produces
Linux amd64 and arm64 package archives with checksums. Extract the archive for the
server's architecture, then install the `tiktok-tts` directory as described below.

For a source build, requirements: Go 1.26 or later and Python 3. Build on a machine matching the
server's operating system and architecture:

```sh
python3 package.py
```

The output directory contains:

```text
dist/tiktok-tts/
  manifest.json
  extension
```

Copy that directory to `HAUNT_DATA_DIR/server-extensions/packages/tiktok-tts/`
on the server, preserving executable permissions. For containers, install it in
the mounted data directory; the executable is statically built with CGO disabled.
Open **Server settings → Server extensions → Reload installed extensions**, or
restart Haunt. **TikTok TTS** then appears as a server-provided settings page.
Updating the package uses the same copy-and-reload sequence. Remove the package
directory and reload to uninstall; saved configuration remains available for a
later reinstall.

## Configure and use

1. Open **Server settings → TikTok TTS**.
2. Choose an API region and enter a TikTok `sessionid` token, a 32-character hex
   value. Save configuration. A blank credential field preserves the saved value.
3. Select a voice and use **Test and play speech**. Tests work while disabled.
4. Enable the extension and save to make it available to signed-in clients.
5. In **Account settings → Text to speech**, choose **TikTok TTS** and a voice.

Haunt stores configuration encrypted on the server and does not return credential
values to clients. The manifest supplies the region choices, credential field,
default voice and 42 grouped voice choices. Personal playback preferences remain
scoped to each server/account.

## Extension interface

This package implements Haunt's [server extension protocol](https://github.com/example-git/haunt-server/blob/main/docs/server-extensions.md#package-contract).
Its manifest supplies the configuration fields and voice catalog; the executable
uses the configured `endpoint` and `session_token` to generate MP3 speech.

The provider accepts up to 500 UTF-8 bytes, splits them into chunks of at most
300 bytes, and uses TikTok's mobile request headers. It accepts only the two
origins declared in the manifest, refuses redirects, bounds response sizes and
uses an eight-second timeout for each upstream request.

## Development

`provider.go` implements TikTok synthesis; `extension.go` implements the host
protocol; `cmd/haunt-tiktok-tts` is the executable entry point. The package can also
be used as a Go library through `Synthesize`.

```sh
go test ./...
```

Tests simulate TikTok HTTP responses and exercise the same protocol handler as
the executable. Live voice availability and credentials can be checked with the
server's **Test and play speech** control.
