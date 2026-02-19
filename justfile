default:
    @just --list

build:
    npx tsc

test-bats: build
    just zz-tests_bats/test

test: test-bats
