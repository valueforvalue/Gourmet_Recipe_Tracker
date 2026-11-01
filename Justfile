# Release version source of truth (stamped into the exe at build time).
release_version := `cat VERSION`

# Compile the Go server into morris-tracker.exe, stamping the version.
build:
    go build -ldflags "-X main.Version={{release_version}}" -o morris-tracker.exe .

# Run the Go test suite.
test:
    go test ./...

# Rebuild the embedded Elm SPA (web/js/elm.js) from web/src/Main.elm.
web:
    cd web && elm make src/Main.elm --output=js/elm.js

# Run the server on :8080 (reads config.json).
run:
    go run -ldflags "-X main.Version={{release_version}}" .

# Print the release version (from VERSION).
version:
    @echo {{release_version}}

# Build the Android release APK.
apk:
    cd flutter_app && flutter build apk --release

# Run the Flutter test suite.
flutter-test:
    cd flutter_app && flutter test

# Run the Flutter app.
flutter-run:
    cd flutter_app && flutter run

# Build, test, and rebuild the Elm bundle (full verification).
check: build test web

# Package server exe + Android APK + runner + VERSION into a versioned zip.
release: build apk
    rm -rf release recipe-tracker-v{{release_version}}.zip
    mkdir -p release
    cp morris-tracker.exe release/
    cp flutter_app/build/app/outputs/flutter-apk/app-release.apk release/morrisrecipetracker.apk
    cp run_tracker.bat release/
    cp config.json.example release/config.json
    cp README.md release/README.md
    echo "{{release_version}}" > release/VERSION
    powershell -NoProfile -Command "Compress-Archive -Path 'release/*' -DestinationPath 'recipe-tracker-v{{release_version}}.zip' -Force"
    rm -rf release

# Default: build the server.
default: build
