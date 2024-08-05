ARG NODE_IMAGE=node:22-alpine
ARG  BUILDER_IMAGE=golang:1.22-alpine
ARG  DISTROLESS_IMAGE=gcr.io/distroless/base
############################
# STEP 1 build tailwindcss
############################
FROM ${NODE_IMAGE} as tailwind
WORKDIR /app

RUN npm i -g pnpm
COPY tailwindcss/package.json tailwindcss/pnpm-lock.yaml ./
RUN pnpm install

COPY tailwindcss/ ./
# We need to copy the templates to build the css
COPY internal/server/templates/ /internal/server/templates/
RUN pnpm run build-css-prod

############################
# STEP 2 build executable binary
############################
FROM ${BUILDER_IMAGE} as builder

# Ensure ca-certficates are up to date
RUN update-ca-certificates

# Set the working directory to the root of your Go module
WORKDIR /app

# Add cache for faster builds
ENV GOCACHE=$HOME/.cache/go-build
RUN --mount=type=cache,target=$GOCACHE

# use modules
COPY go.mod .

RUN go mod download && go mod verify

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /app/steam-avatars ./cmd/steam-avatars

############################
# STEP 3 build a small image
############################
# using base nonroot image
# user:group is nobody:nobody, uid:gid = 65534:65534
FROM ${DISTROLESS_IMAGE}

# Copy our static executable
COPY --from=builder /app/steam-avatars /steam-avatars
COPY --from=builder /app/database.sql /database.sql
COPY --from=builder /app/assets /assets
COPY --from=tailwind /assets/main.css /assets/main.css

EXPOSE 8080
# Run the hello binary.
ENTRYPOINT ["/steam-avatars"]
