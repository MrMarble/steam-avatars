default: setup

setup:
    go install github.com/a-h/templ/cmd/templ@latest
    go install github.com/air-verse/air@latest
    go mod tidy

generate:
    go generate ./...

build:
    go build cmd/steam-avatars

run: generate
    go run ./cmd/steam-avatars

tailwind:
    cd tailwindcss && pnpm build-css-prod

tailwind-watch:
    cd tailwindcss && pnpm watch-css

watch: 
    air
