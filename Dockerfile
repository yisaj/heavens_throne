# build dependencies
FROM golang AS build
RUN mkdir /app
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

# copy source
COPY database database
COPY entities entities
COPY config config
COPY twitlisten twitlisten
COPY twitspeak twitspeak
COPY input input
COPY simulation simulation
COPY main.go main.go

# build the app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# run the app
FROM alpine
RUN apk --no-cache add inkscape
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app

COPY migrations migrations
COPY testcard.png testcard.png
COPY --from=build /app/heavens_throne heavens_throne

EXPOSE 80
EXPOSE 443
ENTRYPOINT "./heavens_throne"
