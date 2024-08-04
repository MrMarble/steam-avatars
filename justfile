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
    just tailwind-watch &
    air
