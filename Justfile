# Compile the Go server into morris-tracker.exe.
build:
    go build -o morris-tracker.exe .

# Run the Go test suite.
test:
    go test ./...

# Rebuild the embedded Elm SPA (web/js/elm.js) from web/src/Main.elm.
web:
    cd web && elm make src/Main.elm --output=js/elm.js

# Run the server on :8080 (reads config.json).
run:
    go run .

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

# Package server exe + Android APK + runner into a release zip (no source).
release: build apk
    rm -rf release recipe-tracker-release.zip
    mkdir -p release
    cp morris-tracker.exe release/
    cp flutter_app/build/app/outputs/flutter-apk/app-release.apk release/morrisrecipetracker.apk
    cp run_tracker.bat release/
    cp config.json.example release/
    powershell -NoProfile -Command "Compress-Archive -Path 'release/*' -DestinationPath 'recipe-tracker-release.zip' -Force"
    rm -rf release

# Default: build the server.
default: build
