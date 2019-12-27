############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR /builder
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /distill -ldflags='-s -w -extldflags "-static"'
############################
# STEP 2 build a small image
############################
FROM scratch
LABEL author="hello@distill.plus"
# Copy our static executable.
COPY --from=builder /distill /
# Copy the temlates folder
# COPY templates /templates
# Copy the settings
COPY configs/settings.docker.yaml /settings.docker.yaml 
# Copy the license
COPY LICENSE /LICENSE
COPY README.md /README.md
# Run the binary.
ENTRYPOINT [ "/distill" ]
CMD [ "start", "-c", "settings.docker.yaml" ]
