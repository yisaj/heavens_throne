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
RUN apk --no-cache add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.10.0/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate.linux-amd64 /usr/bin/migrate 
RUN apk --no-cache add inkscape
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app

COPY migrations migrations
COPY maptemplate.svg maptemplate.svg
COPY LHANDW.TTF /usr/share/fonts/LHANDW.TTF
COPY --from=build /app/heavens_throne heavens_throne

EXPOSE 80
EXPOSE 443
ENTRYPOINT "./heavens_throne"
