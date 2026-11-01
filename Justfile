# Compile the Go server into recipe-tracker.exe.
build:
    go build

# Run the Go test suite.
test:
    go test ./...

# Rebuild the embedded Elm SPA (web/js/elm.js) from web/src/Main.elm.
web:
    cd web && elm make src/Main.elm --output=js/elm.js

# Run the server on :8080 (reads config.json).
run:
    go run .

# Run the Flutter test suite.
flutter-test:
    cd flutter_app && flutter test

# Run the Flutter app.
flutter-run:
    cd flutter_app && flutter run

# Build, test, and rebuild the Elm bundle (full verification).
check: build test web

# Default: build the server.
default: build
