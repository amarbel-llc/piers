default:
    @just --list

build:
    nix build

build-gomod2nix:
    nix develop --command gomod2nix

build-go: build-gomod2nix
    nix develop --command go build -o piers ./cmd/piers

test-go:
    nix develop --command go test ./...

test-bats: build
    nix develop --command just zz-tests_bats/test

test: test-go test-bats

fmt:
    nix develop --command go fmt ./...

deps:
    nix develop --command go mod tidy
    nix develop --command gomod2nix

clean:
    rm -f piers
    rm -rf result
